package paymentaccounts

import (
	"context"
	"fmt"
	"log/slog"

	"cnmt/internal/common"
	"cnmt/internal/common/httpx"
	"cnmt/internal/infra/db"

	"github.com/google/uuid"
)

type Service struct {
	queries *db.Queries
	logger  *slog.Logger
}

func NewService(queries *db.Queries, logger *slog.Logger) *Service {
	return &Service{queries: queries, logger: logger}
}

func (s *Service) ListByCountryID(ctx context.Context, countryID int64) ([]PaymentAccountResponse, error) {
	accounts, err := s.queries.ListActivePaymentAccountsByCountryID(ctx, countryID)
	if err != nil {
		return nil, common.TranslateDBError(err)
	}

	responses := make([]PaymentAccountResponse, 0, len(accounts))
	for _, account := range accounts {
		responses = append(responses, mapListRow(account))
	}
	return responses, nil
}

func (s *Service) Create(ctx context.Context, req CreatePaymentAccountRequest) (PaymentAccountResponse, error) {
	attrs := req.paymentAccountAttrs.normalized()
	if err := attrs.validateMethodFields(req.PaymentMethod); err != nil {
		return PaymentAccountResponse{}, err
	}

	if _, err := s.queries.GetAdminCountryByID(ctx, req.CountryID); err != nil {
		return PaymentAccountResponse{}, common.TranslateDBError(err)
	}

	channel, err := s.validateChannel(ctx, req.CountryID, req.PaymentMethod, attrs.PaymentChannelID)
	if err != nil {
		return PaymentAccountResponse{}, err
	}

	account, err := s.queries.CreatePaymentAccount(ctx, db.CreatePaymentAccountParams{
		CountryID:        req.CountryID,
		PaymentMethod:    req.PaymentMethod,
		Name:             attrs.Name,
		AccountName:      attrs.AccountName,
		AccountNumber:    attrs.AccountNumber,
		PhoneNumber:      attrs.PhoneNumber,
		SortCode:         attrs.SortCode,
		Iban:             attrs.IBAN,
		PaymentChannelID: common.UuidToPgtype(attrs.PaymentChannelID),
		CurrencyCode:     attrs.CurrencyCode,
	})
	if err != nil {
		return PaymentAccountResponse{}, common.TranslateDBError(err)
	}

	channelName := channel.Name
	return mapAccount(account, &channelName), nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdatePaymentAccountRequest) (PaymentAccountResponse, error) {
	existing, err := s.queries.GetPaymentAccountByID(ctx, id)
	if err != nil {
		return PaymentAccountResponse{}, common.TranslateDBError(err)
	}

	attrs := req.paymentAccountAttrs.normalized()
	if err := attrs.validateMethodFields(existing.PaymentMethod); err != nil {
		return PaymentAccountResponse{}, err
	}

	channel, err := s.validateChannel(ctx, existing.CountryID, existing.PaymentMethod, attrs.PaymentChannelID)
	if err != nil {
		return PaymentAccountResponse{}, err
	}

	account, err := s.queries.UpdatePaymentAccount(ctx, db.UpdatePaymentAccountParams{
		ID:               id,
		Name:             attrs.Name,
		AccountName:      attrs.AccountName,
		AccountNumber:    attrs.AccountNumber,
		PhoneNumber:      attrs.PhoneNumber,
		SortCode:         attrs.SortCode,
		Iban:             attrs.IBAN,
		PaymentChannelID: common.UuidToPgtype(attrs.PaymentChannelID),
		CurrencyCode:     attrs.CurrencyCode,
	})
	if err != nil {
		return PaymentAccountResponse{}, common.TranslateDBError(err)
	}

	channelName := channel.Name
	return mapAccount(account, &channelName), nil
}

func (s *Service) Activate(ctx context.Context, id uuid.UUID) (PaymentAccountResponse, error) {
	return s.setActive(ctx, id, true)
}

func (s *Service) Deactivate(ctx context.Context, id uuid.UUID) (PaymentAccountResponse, error) {
	return s.setActive(ctx, id, false)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.queries.DeletePaymentAccount(ctx, id); err != nil {
		return common.TranslateDBError(err)
	}
	return nil
}

func (s *Service) setActive(ctx context.Context, id uuid.UUID, active bool) (PaymentAccountResponse, error) {
	account, err := s.queries.SetPaymentAccountActive(ctx, db.SetPaymentAccountActiveParams{
		ID:       id,
		IsActive: active,
	})
	if err != nil {
		return PaymentAccountResponse{}, common.TranslateDBError(err)
	}

	detail, err := s.queries.GetPaymentAccountByID(ctx, account.ID)
	if err != nil {
		return PaymentAccountResponse{}, common.TranslateDBError(err)
	}
	return mapDetailRow(detail), nil
}

func (s *Service) validateChannel(
	ctx context.Context,
	countryID int64,
	method db.ReceivingMethods,
	channelID uuid.UUID,
) (db.PaymentChannel, error) {
	channel, err := s.queries.GetActivePCByCountryTypeAndID(ctx, db.GetActivePCByCountryTypeAndIDParams{
		CountryID:   countryID,
		ChannelType: method,
		ID:          channelID,
	})
	if err != nil {
		return db.PaymentChannel{}, fmt.Errorf("%w: payment channel not found for country and method", httpx.BadRequestError)
	}
	return channel, nil
}
