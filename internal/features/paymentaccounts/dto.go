package paymentaccounts

import (
	"fmt"
	"strings"
	"time"

	"cnmt/internal/common"
	"cnmt/internal/common/httpx"
	"cnmt/internal/infra/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type paymentAccountAttrs struct {
	Name             string    `json:"name" validate:"required,min=1,max=255"`
	AccountName      string    `json:"account_name" validate:"required,min=1,max=255"`
	AccountNumber    *string   `json:"account_number,omitempty" validate:"omitempty,min=1,max=100"`
	PhoneNumber      *string   `json:"phone_number,omitempty" validate:"omitempty,e164"`
	SortCode         *string   `json:"sort_code,omitempty" validate:"omitempty,min=1,max=32"`
	IBAN             *string   `json:"iban,omitempty" validate:"omitempty,min=1,max=34"`
	PaymentChannelID uuid.UUID `json:"payment_channel_id" validate:"required"`
	CurrencyCode     string    `json:"currency_code" validate:"required,len=3"`
}

type CreatePaymentAccountRequest struct {
	CountryID     int64               `json:"country_id" validate:"required,gt=0"`
	PaymentMethod db.ReceivingMethods `json:"payment_method" validate:"required,oneof=BANK MOBILE_MONEY"`
	paymentAccountAttrs
}

type UpdatePaymentAccountRequest struct {
	paymentAccountAttrs
}

type PaymentAccountResponse struct {
	ID               uuid.UUID           `json:"id"`
	CountryID        int64               `json:"country_id"`
	PaymentMethod    db.ReceivingMethods `json:"payment_method"`
	Name             string              `json:"name"`
	AccountName      string              `json:"account_name"`
	AccountNumber    *string             `json:"account_number,omitempty"`
	PhoneNumber      *string             `json:"phone_number,omitempty"`
	SortCode         *string             `json:"sort_code,omitempty"`
	IBAN             *string             `json:"iban,omitempty"`
	PaymentChannelID *uuid.UUID          `json:"payment_channel_id,omitempty"`
	ChannelName      *string             `json:"channel_name,omitempty"`
	CurrencyCode     string              `json:"currency_code"`
	IsActive         bool                `json:"is_active"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

func (a paymentAccountAttrs) normalized() paymentAccountAttrs {
	a.AccountName = strings.TrimSpace(a.AccountName)
	a.Name = strings.TrimSpace(a.Name)
	a.CurrencyCode = strings.ToUpper(strings.TrimSpace(a.CurrencyCode))
	a.AccountNumber = trimPtr(a.AccountNumber)
	a.PhoneNumber = trimPtr(a.PhoneNumber)
	a.SortCode = trimPtr(a.SortCode)
	a.IBAN = trimPtr(a.IBAN)
	return a
}

func (a paymentAccountAttrs) validateMethodFields(method db.ReceivingMethods) error {
	switch method {
	case db.ReceivingMethodsBANK:
		if a.AccountNumber == nil {
			return fmt.Errorf("%w: account_number is required for bank accounts", httpx.BadRequestError)
		}
		if a.PhoneNumber != nil {
			return fmt.Errorf("%w: phone_number is not allowed for bank accounts", httpx.BadRequestError)
		}
	case db.ReceivingMethodsMOBILEMONEY:
		if a.PhoneNumber == nil {
			return fmt.Errorf("%w: phone_number is required for mobile money accounts", httpx.BadRequestError)
		}
		if a.AccountNumber != nil {
			return fmt.Errorf("%w: account_number is not allowed for mobile money accounts", httpx.BadRequestError)
		}
		if a.SortCode != nil {
			return fmt.Errorf("%w: sort_code is not allowed for mobile money accounts", httpx.BadRequestError)
		}
		if a.IBAN != nil {
			return fmt.Errorf("%w: iban is not allowed for mobile money accounts", httpx.BadRequestError)
		}
	default:
		return fmt.Errorf("%w: unsupported payment method", httpx.BadRequestError)
	}
	return nil
}

func mapListRow(row db.ListActivePaymentAccountsByCountryIDRow) PaymentAccountResponse {
	return PaymentAccountResponse{
		ID:            row.ID,
		CountryID:     row.CountryID,
		PaymentMethod: row.PaymentMethod,
		Name:          row.Name,
		AccountName:   row.AccountName,
		AccountNumber: row.AccountNumber,
		PhoneNumber:   row.PhoneNumber,
		SortCode:      row.SortCode,
		IBAN:          row.Iban,
		ChannelName:   row.ChannelName,
		CurrencyCode:  row.CurrencyCode,
		IsActive:      true,
	}
}

func mapAdminListRow(row db.ListPaymentAccountsRow) PaymentAccountResponse {
	return PaymentAccountResponse{
		ID:               row.ID,
		CountryID:        row.CountryID,
		PaymentMethod:    row.PaymentMethod,
		Name:             row.Name,
		AccountName:      row.AccountName,
		AccountNumber:    row.AccountNumber,
		PhoneNumber:      row.PhoneNumber,
		SortCode:         row.SortCode,
		IBAN:             row.Iban,
		PaymentChannelID: pgUuidPtr(row.PaymentChannelID),
		ChannelName:      row.ChannelName,
		CurrencyCode:     row.CurrencyCode,
		IsActive:         row.IsActive,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}

func mapDetailRow(row db.GetPaymentAccountByIDRow) PaymentAccountResponse {
	return PaymentAccountResponse{
		ID:               row.ID,
		CountryID:        row.CountryID,
		PaymentMethod:    row.PaymentMethod,
		Name:             row.Name,
		AccountName:      row.AccountName,
		AccountNumber:    row.AccountNumber,
		PhoneNumber:      row.PhoneNumber,
		SortCode:         row.SortCode,
		IBAN:             row.Iban,
		PaymentChannelID: pgUuidPtr(row.PaymentChannelID),
		ChannelName:      row.ChannelName,
		CurrencyCode:     row.CurrencyCode,
		IsActive:         row.IsActive,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}

func mapAccount(account db.PaymentAccount, channelName *string) PaymentAccountResponse {
	return PaymentAccountResponse{
		ID:               account.ID,
		CountryID:        account.CountryID,
		PaymentMethod:    account.PaymentMethod,
		Name:             account.Name,
		AccountName:      account.AccountName,
		AccountNumber:    account.AccountNumber,
		PhoneNumber:      account.PhoneNumber,
		SortCode:         account.SortCode,
		IBAN:             account.Iban,
		PaymentChannelID: pgUuidPtr(account.PaymentChannelID),
		ChannelName:      channelName,
		CurrencyCode:     account.CurrencyCode,
		IsActive:         account.IsActive,
		CreatedAt:        account.CreatedAt,
		UpdatedAt:        account.UpdatedAt,
	}
}

func trimPtr(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func pgUuidPtr(id pgtype.UUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}
	u := common.PgUuidToUuid(id)
	return &u
}
