package transfers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"cnmt/internal/common"
	"cnmt/internal/common/httpx"
	"cnmt/internal/infra/db"
	"cnmt/internal/infra/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type Service struct {
	db      *pgxpool.Pool
	queries *db.Queries
	objStorage *storage.ObjStorage
	logger *slog.Logger
}


func NewService(db *pgxpool.Pool, queries *db.Queries, objStorage *storage.ObjStorage, logger *slog.Logger) *Service {
	return &Service{db: db, queries: queries, objStorage: objStorage, logger: logger}
}

func (s *Service) CreateTransfer(ctx context.Context, body createTransferRequest, idemKey string) (createTransferResponse, error) {
	if idemKey == "" {
		return createTransferResponse{}, fmt.Errorf("%w: idempotency key is required", httpx.BadRequestError)
	}
	actorID := body.SenderPhone
	reqHash := hashCreateTransferRequest(body)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		s.logger.Error("failed to begin transaction", "error", err)
		return createTransferResponse{}, common.TranslateDBError(err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	_, err = qtx.InsertIdempotencyKey(ctx, db.InsertIdempotencyKeyParams{
		Key:         idemKey,
		ActorID:     actorID,
		RequestHash: reqHash,
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
	})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			s.logger.Error("failed to insert idempotency key", "error", err)
			return createTransferResponse{}, common.TranslateDBError(err)
		}

		existing, getErr := qtx.GetIdempotencyKey(ctx, db.GetIdempotencyKeyParams{
			ActorID: actorID,
			Key:     idemKey,
		})
		if getErr != nil {
			s.logger.Error("failed to get idempotency key", "error", getErr)
			return createTransferResponse{}, common.TranslateDBError(getErr)
		}
		if existing.RequestHash != reqHash {
			return createTransferResponse{}, fmt.Errorf("%w: idempotency key reused with a different request", httpx.ConflictError)
		}
		if existing.Status == "completed" {
			var resp createTransferResponse
			if err := json.Unmarshal(existing.ResponseBody, &resp); err != nil {
				s.logger.Error("failed to unmarshal idempotency key response body", "error", err)
				return createTransferResponse{}, fmt.Errorf("%w", httpx.InternalServerError)
			}
			return resp, nil
		}
		return createTransferResponse{}, fmt.Errorf("%w: request with this idempotency key is already in progress", httpx.ConflictError)
	}

	resp, err := s.createTransfer(ctx, qtx, body)
	if err != nil {
		s.logger.Error("failed to create transfer", "error", err)
		return createTransferResponse{}, err
	}

	raw, err := json.Marshal(resp)
	if err != nil {
		s.logger.Error("failed to marshal transfer response", "error", err)
		return createTransferResponse{}, fmt.Errorf("%w", httpx.InternalServerError)
	}
	code := int32(http.StatusCreated)
	if err := qtx.CompleteIdempotencyKey(ctx, db.CompleteIdempotencyKeyParams{
		ActorID:      actorID,
		Key:          idemKey,
		ResponseCode: &code,
		ResponseBody: raw,
		TransferID:   common.UuidToPgtype(resp.TransferID),
	}); err != nil {
		s.logger.Error("failed to complete idempotency key", "error", err)
		return createTransferResponse{}, common.TranslateDBError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Error("failed to commit transaction", "error", err)
		return createTransferResponse{}, common.TranslateDBError(err)
	}

	return resp, nil
}

func (s *Service) createTransfer(ctx context.Context, q *db.Queries, body createTransferRequest) (createTransferResponse, error) {
	if body.SourceCountryID == body.DestinationCountryID {
		return createTransferResponse{}, fmt.Errorf("%w: source and destination countries must be different", httpx.BadRequestError)
	}

	route, err := q.GetActiveRouteByCountries(ctx, db.GetActiveRouteByCountriesParams{
		SourceCountryID:      body.SourceCountryID,
		DestinationCountryID: body.DestinationCountryID,
	})
	if err != nil {
		s.logger.Error("failed to get active route by countries", "error", err)
		return createTransferResponse{}, common.TranslateDBError(err)
	}

	amtPaid, err := decimal.NewFromString(body.AmountSent)
	if err != nil {	
		s.logger.Error("failed to convert amount sent to decimal", "error", err)
		return createTransferResponse{}, fmt.Errorf("%w: amount sent is not valid", httpx.BadRequestError)
	}
	amtPaid = common.RoundMoney(amtPaid)

	if err := validateAmountPaid(amtPaid, route.MinTransferAmount, route.MaxTransferAmount); err != nil {
		s.logger.Error("failed to validate amount paid", "error", err)
		return createTransferResponse{}, err
	}

	rate, err := common.PgNumericToDecimal(route.DefaultExchangeRate)
	if err != nil || !rate.IsPositive() {
		s.logger.Error("failed to convert exchange rate to decimal", "error", err)
		return createTransferResponse{}, fmt.Errorf("%w: exchange rate is not configured", httpx.BadRequestError)
	}

	calculatedFee, err := calculateFee(amtPaid, route.FeeType, route.Fee)
	if err != nil {
		s.logger.Error("failed to calculate fee", "error", err)
		return createTransferResponse{}, err
	}

	amtToReceive := common.RoundMoney(amtPaid.Mul(rate).Sub(calculatedFee))
	if !amtToReceive.IsPositive() {
		return createTransferResponse{}, fmt.Errorf("%w: amount received must be greater than zero after fees", httpx.BadRequestError)
	}

	networkID, bankID, mobileNumber, bankAccount, err := receivingDetails(body.Recipient)
	if err != nil {
		s.logger.Error("failed to get receiving details", "error", err)
		return createTransferResponse{}, err
	}

	if err := validatePaymentChannel(ctx, q, body.DestinationCountryID, body.Recipient.ReceivingMethod, networkID, bankID); err != nil {
		s.logger.Error("failed to validate payment channel", "error", err)
		return createTransferResponse{}, err
	}

	feeNumeric, err := common.DecimalToPgNumeric(calculatedFee)
	if err != nil {
		s.logger.Error("failed to convert fee to pgnumeric", "error", err)
		return createTransferResponse{}, fmt.Errorf("%w", httpx.InternalServerError)
	}
	sentNumeric, err := common.DecimalToPgNumeric(amtPaid)
	if err != nil {
		s.logger.Error("failed to convert amount sent to pgnumeric", "error", err)
		return createTransferResponse{}, fmt.Errorf("%w", httpx.InternalServerError)
	}
	receivedNumeric, err := common.DecimalToPgNumeric(amtToReceive)
	if err != nil {
		s.logger.Error("failed to convert amount received to pgnumeric", "error", err)
		return createTransferResponse{}, fmt.Errorf("%w", httpx.InternalServerError)
	}

	reference := common.GenerateReference()
	if reference == "" {
		return createTransferResponse{}, fmt.Errorf("%w", httpx.InternalServerError)
	}

	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	transferID, err := q.CreateTransfer(ctx, db.CreateTransferParams{
		Reference:                  reference,
		RouteID:                    route.ID,
		Status:                     db.TransferStatusPENDINGPAYMENT,
		SenderPhone:                body.SenderPhone,
		ReceivingAccountName:       body.Recipient.RecipientName,
		ReceivingMobileMoneyNumber: mobileNumber,
		ReceivingMethod:            body.Recipient.ReceivingMethod,
		ReceivingMoneyNetworkID:    networkID,
		ReceivingBankID:            bankID,
		ReceivingBankAccount:       bankAccount,
		PaymentProofKey:            nil,
		ExchangeRate:               route.DefaultExchangeRate,
		Fee:                        feeNumeric,
		AmountSent:                 sentNumeric,
		AmountReceived:             receivedNumeric,
		Notes:                      body.Notes,
		ExpiresAt:                  expiresAt,
	})
	if err != nil {
		s.logger.Error("failed to create transfer", "error", err)
		return createTransferResponse{}, common.TranslateDBError(err)
	}

	return createTransferResponse{
		TransferID: transferID,
		Reference:  reference,
		ExpiresIn:  expiresAt.Unix(),
	}, nil
}

func receivingDetails(recipient *recipientDTO) (networkID, bankID pgtype.UUID, mobileNumber, bankAccount *string, err error) {
	if recipient == nil {
		return pgtype.UUID{}, pgtype.UUID{}, nil, nil, fmt.Errorf("%w: recipient is required", httpx.BadRequestError)
	}

	switch recipient.ReceivingMethod {
	case db.ReceivingMethodsBANK:
		if recipient.BankID == nil {
			return pgtype.UUID{}, pgtype.UUID{}, nil, nil, fmt.Errorf("%w: bank is required", httpx.BadRequestError)
		}
		return pgtype.UUID{}, common.UuidPtrToPgtype(recipient.BankID), nil, recipient.AccountNumber, nil
	case db.ReceivingMethodsMOBILEMONEY:
		if recipient.ReceivingNetworkID == nil {
			return pgtype.UUID{}, pgtype.UUID{}, nil, nil, fmt.Errorf("%w: mobile money network is required", httpx.BadRequestError)
		}
		return common.UuidPtrToPgtype(recipient.ReceivingNetworkID), pgtype.UUID{}, recipient.RecipientPhone, nil, nil
	default:
		return pgtype.UUID{}, pgtype.UUID{}, nil, nil, fmt.Errorf("%w: unsupported receiving method", httpx.BadRequestError)
	}
}

func validatePaymentChannel(ctx context.Context, q *db.Queries, countryID int64, method db.ReceivingMethods, networkID, bankID pgtype.UUID) error {
	var channelID uuid.UUID
	switch method {
	case db.ReceivingMethodsBANK:
		if !bankID.Valid {
			return fmt.Errorf("%w: bank not found", httpx.BadRequestError)
		}
		channelID = bankID.Bytes
	case db.ReceivingMethodsMOBILEMONEY:
		if !networkID.Valid {
			return fmt.Errorf("%w: mobile money network not found", httpx.BadRequestError)
		}
		channelID = networkID.Bytes
	default:
		return fmt.Errorf("%w: unsupported receiving method", httpx.BadRequestError)
	}

	pch, err := q.GetActivePCByCountryTypeAndID(ctx, db.GetActivePCByCountryTypeAndIDParams{
		CountryID:   countryID,
		ChannelType: method,
		ID:          channelID,
	})
	if err != nil {
		return common.TranslateDBError(err)
	}
	if pch.ID == uuid.Nil {
		return fmt.Errorf("%w: payment channel not found", httpx.NotFoundError)
	}
	return nil
}

func validateAmountPaid(amountPaid decimal.Decimal, minTransferAmount, maxTransferAmount pgtype.Numeric) error {
	if !amountPaid.IsPositive() {
		return fmt.Errorf("%w: amount sent must be greater than zero", httpx.BadRequestError)
	}

	if minTransferAmount.Valid {
		minAmount, err := common.PgNumericToDecimal(minTransferAmount)
		if err != nil {
			return fmt.Errorf("%w: route minimum amount is invalid", httpx.BadRequestError)
		}
		if amountPaid.LessThan(minAmount) {
			return fmt.Errorf("%w: amount sent is less than the minimum transfer amount", httpx.BadRequestError)
		}
	}

	if maxTransferAmount.Valid {
		maxAmount, err := common.PgNumericToDecimal(maxTransferAmount)
		if err != nil {
			return fmt.Errorf("%w: route maximum amount is invalid", httpx.BadRequestError)
		}
		if !maxAmount.IsZero() && amountPaid.GreaterThan(maxAmount) {
			return fmt.Errorf("%w: amount sent is greater than the maximum transfer amount", httpx.BadRequestError)
		}
	}

	return nil
}

func calculateFee(amountPaid decimal.Decimal, feeType db.FeeType, fee pgtype.Numeric) (decimal.Decimal, error) {
	feeValue, err := common.PgNumericToDecimal(fee)
	if err != nil {
		return decimal.Zero, fmt.Errorf("%w: fee is not configured", httpx.BadRequestError)
	}
	if feeValue.IsNegative() {
		return decimal.Zero, fmt.Errorf("%w: fee is not valid", httpx.BadRequestError)
	}

	switch feeType {
	case db.FeeTypeFixed:
		return common.RoundMoney(feeValue), nil
	case db.FeeTypePercentage:
		if feeValue.GreaterThan(decimal.NewFromInt(100)) {
			return decimal.Zero, fmt.Errorf("%w: fee is not valid", httpx.BadRequestError)
		}
		return common.RoundMoney(amountPaid.Mul(feeValue).Div(decimal.NewFromInt(100))), nil
	default:
		return decimal.Zero, fmt.Errorf("%w: unsupported fee type", httpx.BadRequestError)
	}
}

func hashCreateTransferRequest(body createTransferRequest) string {
	payload, _ := json.Marshal(body)
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}


func (s *Service) GetByReference(ctx context.Context, reference string) (getTransferResponse, error) {
	if reference == "" {
		return getTransferResponse{}, fmt.Errorf("%w: reference is required", httpx.BadRequestError)
	}

	transfer, err := s.queries.GetTransferByReference(ctx, reference)
	if err != nil {
		s.logger.Error("failed to get transfer by reference", "error", err)
		return getTransferResponse{}, common.TranslateDBError(err)
	}

	transferDTO, err := mapTransferToDTO(transfer)
	if err != nil {
		s.logger.Error("failed to map transfer to dto", "error", err)
		return getTransferResponse{}, fmt.Errorf("%w", httpx.InternalServerError)
	}
	return transferDTO, nil
}

func (s *Service) CreatePaymentProofSignedUrl(ctx context.Context, reference string) (createPaymentProofSignedUrlResponse, error) {
	if reference == "" {
		return createPaymentProofSignedUrlResponse{}, fmt.Errorf("%w: reference is required", httpx.BadRequestError)
	}

	transfer, err := s.queries.GetTransferByReference(ctx, reference)
	if err != nil {
		return createPaymentProofSignedUrlResponse{}, common.TranslateDBError(err)
	}

	if transfer.Status != db.TransferStatusPENDINGPAYMENT {
		return createPaymentProofSignedUrlResponse{}, fmt.Errorf("%w: transfer is not pending payment", httpx.BadRequestError)
	}

	key := common.GenerateAssetKey("transfers", transfer.Reference, "payment-proof", storage.ContentTypeImage)
	signedUrl, err := s.objStorage.GetPresignedUrl(ctx, key, storage.ContentTypeImage)
	if err != nil {
		return createPaymentProofSignedUrlResponse{}, common.TranslateDBError(err)
	}

	return createPaymentProofSignedUrlResponse{
		SignedURL: signedUrl,
		Key: key,
	}, nil
}

func (s *Service) ConfirmPaymentProof(ctx context.Context, body confirmPaymentProofRequest) error {
	
	transfer, err := s.queries.GetTransferByReference(ctx, body.Reference)
	if err != nil {
		return common.TranslateDBError(err)
	}

	if transfer.Status != db.TransferStatusPENDINGPAYMENT {
		return fmt.Errorf("%w: transfer is not pending payment", httpx.BadRequestError)
	}
	
	if transfer.PaymentProofKey != nil && *transfer.PaymentProofKey != body.Key {
		return fmt.Errorf("%w: payment proof key is incorrect", httpx.BadRequestError)
	}

	exists, err := s.objStorage.DoesObjectExist(ctx, body.Key)
	if err != nil {
		return fmt.Errorf("%w: Failed to verify payment proof", httpx.InternalServerError)
	}
	if !exists {
		return fmt.Errorf("%w: Payment proof has not been uploaded", httpx.BadRequestError)
	}

	if err := s.queries.SetPaymentProofKey(ctx, db.SetPaymentProofKeyParams{
		Reference: body.Reference,
		PaymentProofKey: &body.Key,
	}); err != nil {
		return common.TranslateDBError(err)
	}

	return nil
}