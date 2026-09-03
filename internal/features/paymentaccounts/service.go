package paymentaccounts

import (
	"context"
	"fmt"
	"log/slog"

	"cnmt/internal/common"
	"cnmt/internal/common/httpx"
	"cnmt/internal/infra/db"
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
	req = req.normalized()
	if err := req.validateMethodFields(); err != nil {
		return PaymentAccountResponse{}, err
	}

	if _, err := s.queries.GetAdminCountryByID(ctx, req.CountryID); err != nil {
		return PaymentAccountResponse{}, common.TranslateDBError(err)
	}

	channel, err := s.queries.GetActivePCByCountryTypeAndID(ctx, db.GetActivePCByCountryTypeAndIDParams{
		CountryID:   req.CountryID,
		ChannelType: req.PaymentMethod,
		ID:          req.PaymentChannelID,
	})
	if err != nil {
		return PaymentAccountResponse{}, fmt.Errorf("%w: payment channel not found for country and method", httpx.BadRequestError)
	}

	account, err := s.queries.CreatePaymentAccount(ctx, db.CreatePaymentAccountParams{
		CountryID:        req.CountryID,
		PaymentMethod:    req.PaymentMethod,
		Name:             req.Name,
		AccountName:      req.AccountName,
		AccountNumber:    req.AccountNumber,
		PhoneNumber:      req.PhoneNumber,
		SortCode:         req.SortCode,
		Iban:             req.IBAN,
		PaymentChannelID: common.UuidToPgtype(req.PaymentChannelID),
		CurrencyCode:     req.CurrencyCode,
	})
	if err != nil {
		return PaymentAccountResponse{}, common.TranslateDBError(err)
	}

	return mapCreated(account, channel.Name), nil
}
