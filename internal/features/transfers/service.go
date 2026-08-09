package transfers

import (
	"cnmt/internal/common"
	"cnmt/internal/infra/db"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type Service struct {
	db *pgxpool.Pool
	queries *db.Queries
}

func NewService(db *pgxpool.Pool, queries *db.Queries) *Service {
	return &Service{db: db, queries: queries}
}

func (s *Service) CreateTransfer(ctx context.Context, body createTransferRequest) (createTransferResponse, error) {
	// validate route
	route, err := s.queries.GetActiveRouteByCountries(ctx, db.GetActiveRouteByCountriesParams{
		SourceCountryID: body.SourceCountryID,
		DestinationCountryID: body.DestinationCountryID,
	})
	if err != nil {
		return createTransferResponse{}, common.TranslateDBError(err)
	}
	if route.ID == uuid.Nil {
		return createTransferResponse{}, errors.New("route not found")
	}

	amtPaid,amountPaidErr := decimal.NewFromString(body.AmountSent)
	if amountPaidErr != nil {
		return createTransferResponse{}, errors.New("amount sent is not valid")
	}
	
	if err := validateAmountPaid(amtPaid, route.MinTransferAmount, route.MaxTransferAmount); err != nil {
		return createTransferResponse{}, err
	}

	
	// validate amount received
	decimalExchangeRate := common.ConvertPgNumericToDecimal(route.DefaultExchangeRate)
	

	calculatedFee, calculatedFeeErr := calculateFee(amtPaid, route.FeeType, route.Fee)
	if calculatedFeeErr != nil {
		return createTransferResponse{}, calculatedFeeErr
	}

	amtToReceive := amtPaid.Mul(decimalExchangeRate)
	amtToReceive = amtToReceive.Sub(calculatedFee)

	// does network/bank exist?
	if err := s.validateNetwork(ctx, body.DestinationCountryID, body.Recipient.ReceivingMethod, body.Recipient.ReceivingNetworkID, body.Recipient.BankID); err != nil {
		return createTransferResponse{}, err
	}

	reference := common.GenerateReference()

	expiresIn := time.Now().Add(time.Hour * 24) // 24 hours

	params := db.CreateTransferParams{
		Reference: reference,
    	RouteID: route.ID,
		Status: db.TransferStatusPENDINGPAYMENT,
		SenderPhone: body.SenderPhone,
		ReceivingAccountName: body.Recipient.RecipientName,
		ReceivingMobileMoneyNumber: body.Recipient.RecipientPhone,
		ReceivingMethod: body.Recipient.ReceivingMethod,
		ReceivingMoneyNetworkID: common.UuidPtrToPgtype(body.Recipient.ReceivingNetworkID),
		ReceivingBankID:         common.UuidPtrToPgtype(body.Recipient.BankID),
		ReceivingBankAccount: body.Recipient.AccountNumber,
		PaymentProofKey: nil,
		ExchangeRate: route.DefaultExchangeRate,
		Fee: calculatedFee,
		AmountSent: amtPaid,
		AmountReceived: amtToReceive,
		Notes: body.Notes,
		ExpiresAt: expiresIn,
	}
	

	transfer, err := s.queries.CreateTransfer(ctx, params)
	if err != nil {
		return createTransferResponse{}, common.TranslateDBError(err)
	}
	return createTransferResponse{
		TransferID: transfer,
		Reference: reference,
		ExpiresIn: expiresIn.Unix(),
	}, nil
}

func (s *Service) validateNetwork(ctx context.Context, recipientCountryID int64, receivingMethod db.ReceivingMethods, receivingNetworkID,bankID *uuid.UUID) error {
	var moneyNetworkId uuid.UUID
	switch receivingMethod {
	case db.ReceivingMethodsBANK:
		if bankID == nil {
			return errors.New("bank not found")
		}
		moneyNetworkId = *bankID
	case db.ReceivingMethodsMOBILEMONEY:
		if receivingNetworkID == nil {
			return errors.New("mobile money network not found")
		}
		moneyNetworkId = *receivingNetworkID
	}
	if moneyNetworkId == uuid.Nil {
		return errors.New("Payment channel not found")
	}
	
	pch, err := s.queries.GetActivePCByCountryTypeAndID(ctx, db.GetActivePCByCountryTypeAndIDParams{
		CountryID: recipientCountryID,
		ChannelType: receivingMethod,
		ID: moneyNetworkId,
	})
	if err != nil {
		return err
	}
	if pch.ID == uuid.Nil {
		return errors.New("Payment channel not found")
	}
	return nil
}


func validateAmountPaid(amountPaid decimal.Decimal, minTransferAmount, maxTransferAmount pgtype.Numeric) error {
	
	minTransferAmountDecimal := common.ConvertPgNumericToDecimal(minTransferAmount)
	maxTransferAmountDecimal := common.ConvertPgNumericToDecimal(maxTransferAmount)
	
	if !minTransferAmountDecimal.IsZero() && amountPaid.LessThan(minTransferAmountDecimal) {
		return errors.New("amount paid is less than the minimum transfer amount required to complete the transfer")
	}
	if !maxTransferAmountDecimal.IsZero() && amountPaid.GreaterThan(maxTransferAmountDecimal) {
		return errors.New("amount paid is greater than the maximum transfer amount allowed to complete the transfer")
	}
	
	return nil
}

func calculateFee(amountPaid decimal.Decimal, feeType db.FeeType, fee pgtype.Numeric) (decimal.Decimal, error) {
	var decimalFee decimal.Decimal
	if feeType == db.FeeTypeFixed {
		decimalFee = common.ConvertPgNumericToDecimal(fee)
	} else {
		feeDecimal := common.ConvertPgNumericToDecimal(fee)
		if feeDecimal.LessThan(decimal.Zero) || feeDecimal.GreaterThan(decimal.NewFromInt(100)) {
			return decimal.Zero, errors.New("fee is not valid")
		}
		decimalFee = feeDecimal.Div(decimal.NewFromInt(100))
		decimalFee = decimalFee.Mul(amountPaid)
	}
	return decimalFee, nil
}