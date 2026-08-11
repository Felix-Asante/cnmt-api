package transfers

import (
	"time"

	"cnmt/internal/common"
	"cnmt/internal/infra/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type createTransferRequest struct {
	SourceCountryID      int64 `json:"source_country_id" validate:"gt=0"`
	DestinationCountryID int64 `json:"destination_country_id" validate:"gt=0"`

	AmountSent string `json:"amount_sent" validate:"required,numeric"`

	SenderPhone string `json:"sender_phone" validate:"required,e164"`

	Recipient *recipientDTO `json:"recipient" validate:"required"`

	Notes *string `json:"notes,omitempty" validate:"omitempty,max=500"`
}

type recipientDTO struct {
	RecipientName  string  `json:"recipient_name" validate:"required,min=2,max=100"`
	RecipientPhone *string `json:"recipient_phone,omitempty" validate:"omitempty,e164,required_if=ReceivingMethod MOBILE_MONEY"`

	ReceivingMethod db.ReceivingMethods `json:"receiving_method" validate:"required,oneof=BANK MOBILE_MONEY"`

	ReceivingNetworkID *uuid.UUID `json:"receiving_network_id,omitempty" validate:"omitempty,uuid,required_if=ReceivingMethod MOBILE_MONEY"`

	BankID *uuid.UUID `json:"bank_id,omitempty" validate:"omitempty,uuid,required_if=ReceivingMethod BANK"`

	AccountNumber *string `json:"account_number,omitempty" validate:"omitempty,alphanum,required_if=ReceivingMethod BANK"`
}

type createTransferResponse struct {
	TransferID uuid.UUID `json:"transfer_id"`
	Reference  string    `json:"reference"`
	ExpiresIn  int64     `json:"expires_in"`
}

type countryDTO struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Flag           string `json:"flag,omitempty"`
	CurrencyCode   string `json:"currency_code,omitempty"`
	CurrencySymbol string `json:"currency_symbol,omitempty"`
}

type routeDTO struct {
	SourceCountry      countryDTO `json:"source_country"`
	DestinationCountry countryDTO `json:"destination_country"`
}

type recipientViewDTO struct {
	Name            string              `json:"name"`
	Phone           *string             `json:"phone,omitempty"`
	ReceivingMethod db.ReceivingMethods `json:"receiving_method"`
	NetworkName     *string             `json:"network_name,omitempty"`
	BankName        *string             `json:"bank_name,omitempty"`
	AccountNumber   *string             `json:"account_number,omitempty"`
}


type getTransferResponse struct {
	Reference           string                  `json:"reference"`
	Status              db.TransferStatus       `json:"status"`
	Route               routeDTO                `json:"route"`
	AmountSent          decimal.Decimal         `json:"amount_sent"`
	AmountReceived      decimal.Decimal         `json:"amount_received"`
	ExchangeRate        decimal.Decimal         `json:"exchange_rate"`
	Fee                 decimal.Decimal         `json:"fee"`
	SenderPhone         string                  `json:"sender_phone"`
	PaymentProofKey     *string                 `json:"payment_proof_key,omitempty"`
	Recipient           recipientViewDTO        `json:"recipient"`
	Notes               *string                 `json:"notes,omitempty"`
	ExpiresAt           time.Time               `json:"expires_at"`
	CreatedAt           time.Time               `json:"created_at"`
}


type createPaymentProofSignedUrlRequest struct {
	Reference   string `json:"reference" validate:"required"`
	ContentType string `json:"content_type" validate:"required,oneof=image/jpeg image/jpg image/png"`
}

type createPaymentProofSignedUrlResponse struct {
	SignedURL   string `json:"signed_url"`
	Key         string `json:"key"`
	ContentType string `json:"content_type"`
}

type confirmPaymentProofRequest struct {
	Reference string `json:"reference" validate:"required,min=1,max=100"`
	Key       string `json:"key" validate:"required,min=1"`
}

type getAllTransfersRequest struct {
	SenderPhone    *string              `validate:"omitempty,e164"`
	RecipientPhone *string              `validate:"omitempty,e164"`
	RouteID        *uuid.UUID           `validate:"omitempty,uuid"`
	Reference      *string              `validate:"omitempty,min=1,max=100"`
	Status         *db.TransferStatus   `validate:"omitempty,oneof=PENDING_PAYMENT PAYMENT_RECEIVED VERIFYING PROCESSING COMPLETED FAILED CANCELLED"`
	Page           *int                 `validate:"omitempty,gte=1"`
	Limit          *int                 `validate:"omitempty,gte=1,lte=100"`
}

type getAllTransfersResponse struct {
	Transfers []getTransferResponse `json:"transfers"`
	Total int `json:"total"`
	Page int `json:"page"`
	Limit int `json:"limit"`
}

func mapTransferToDTO(transfer db.GetTransferByReferenceRow) (getTransferResponse, error) {
	return mapTransferRowToDTO(transferRowFromReference(transfer))
}

func mapListTransferToDTO(transfer db.GetAllTransfersRow) (getTransferResponse, error) {
	return mapTransferRowToDTO(transferRowFromList(transfer))
}

type transferRow struct {
	Reference                  string
	Status                     db.TransferStatus
	SourceCountryID            int64
	SourceCountryName          string
	SourceCountryFlag          string
	SourceCurrencyCode         string
	SourceCurrencySymbol       string
	DestinationCountryID       int64
	DestinationCountryName     string
	DestinationCountryFlag     string
	DestinationCurrencyCode    string
	DestinationCurrencySymbol  string
	AmountSent                 pgtype.Numeric
	AmountReceived             pgtype.Numeric
	ExchangeRate               pgtype.Numeric
	Fee                        pgtype.Numeric
	SenderPhone                string
	PaymentProofKey            *string
	ReceivingAccountName       string
	ReceivingMobileMoneyNumber *string
	ReceivingMethod            db.ReceivingMethods
	ReceivingNetworkName       *string
	ReceivingBankName          *string
	ReceivingBankAccount       *string
	Notes                      *string
	ExpiresAt                  time.Time
	CreatedAt                  time.Time
}

func transferRowFromReference(row db.GetTransferByReferenceRow) transferRow {
	return transferRow{
		Reference:                  row.Reference,
		Status:                     row.Status,
		SourceCountryID:            row.SourceCountryID,
		SourceCountryName:          row.SourceCountryName,
		SourceCurrencyCode:         row.SourceCurrencyCode,
		SourceCurrencySymbol:       row.SourceCurrencySymbol,
		DestinationCountryID:       row.DestinationCountryID,
		DestinationCountryName:     row.DestinationCountryName,
		DestinationCurrencyCode:    row.DestinationCurrencyCode,
		DestinationCurrencySymbol:  row.DestinationCurrencySymbol,
		AmountSent:                 row.AmountSent,
		AmountReceived:             row.AmountReceived,
		ExchangeRate:               row.ExchangeRate,
		Fee:                        row.Fee,
		SenderPhone:                row.SenderPhone,
		PaymentProofKey:            row.PaymentProofKey,
		ReceivingAccountName:       row.ReceivingAccountName,
		ReceivingMobileMoneyNumber: row.ReceivingMobileMoneyNumber,
		ReceivingMethod:            row.ReceivingMethod,
		ReceivingNetworkName:       row.ReceivingNetworkName,
		ReceivingBankName:          row.ReceivingBankName,
		ReceivingBankAccount:       row.ReceivingBankAccount,
		Notes:                      row.Notes,
		ExpiresAt:                  row.ExpiresAt,
		CreatedAt:                  row.CreatedAt,
	}
}

func transferRowFromList(row db.GetAllTransfersRow) transferRow {
	return transferRow{
		Reference:                  row.Reference,
		Status:                     row.Status,
		SourceCountryID:            row.SourceCountryID,
		SourceCountryName:          row.SourceCountryName,
		SourceCountryFlag:          row.SourceFlag,
		SourceCurrencySymbol:       row.SourceCurrencySymbol,
		DestinationCountryID:       row.DestinationCountryID,
		DestinationCountryName:     row.DestinationCountryName,
		DestinationCountryFlag:     row.DestinationFlag,
		DestinationCurrencySymbol:  row.DestinationCurrencySymbol,
		AmountSent:                 row.AmountSent,
		AmountReceived:             row.AmountReceived,
		ExchangeRate:               row.ExchangeRate,
		Fee:                        row.Fee,
		SenderPhone:                row.SenderPhone,
		PaymentProofKey:            row.PaymentProofKey,
		ReceivingAccountName:       row.ReceivingAccountName,
		ReceivingMobileMoneyNumber: row.ReceivingMobileMoneyNumber,
		ReceivingMethod:            row.ReceivingMethod,
		ReceivingNetworkName:       row.ReceivingNetworkName,
		ReceivingBankName:          row.ReceivingBankName,
		ReceivingBankAccount:       row.ReceivingBankAccount,
		Notes:                      row.Notes,
		ExpiresAt:                  row.ExpiresAt,
		CreatedAt:                  row.CreatedAt,
	}
}

func mapTransferRowToDTO(transfer transferRow) (getTransferResponse, error) {
	amountSent, err := common.PgNumericToDecimal(transfer.AmountSent)
	if err != nil {
		return getTransferResponse{}, err
	}
	amountReceived, err := common.PgNumericToDecimal(transfer.AmountReceived)
	if err != nil {
		return getTransferResponse{}, err
	}
	exchangeRate, err := common.PgNumericToDecimal(transfer.ExchangeRate)
	if err != nil {
		return getTransferResponse{}, err
	}
	fee, err := common.PgNumericToDecimal(transfer.Fee)
	if err != nil {
		return getTransferResponse{}, err
	}

	resp := getTransferResponse{
		Reference: transfer.Reference,
		Status:    transfer.Status,
		Route: routeDTO{
			SourceCountry: countryDTO{
				ID:             transfer.SourceCountryID,
				Name:           transfer.SourceCountryName,
				Flag:           transfer.SourceCountryFlag,
				CurrencyCode:   transfer.SourceCurrencyCode,
				CurrencySymbol: transfer.SourceCurrencySymbol,
			},
			DestinationCountry: countryDTO{
				ID:             transfer.DestinationCountryID,
				Name:           transfer.DestinationCountryName,
				Flag:           transfer.DestinationCountryFlag,
				CurrencyCode:   transfer.DestinationCurrencyCode,
				CurrencySymbol: transfer.DestinationCurrencySymbol,
			},
		},
		AmountSent:     amountSent,
		AmountReceived: amountReceived,
		ExchangeRate:   exchangeRate,
		Fee:            fee,
		SenderPhone:    transfer.SenderPhone,
		PaymentProofKey: transfer.PaymentProofKey,
		Recipient: recipientViewDTO{
			Name:            transfer.ReceivingAccountName,
			Phone:           transfer.ReceivingMobileMoneyNumber,
			ReceivingMethod: transfer.ReceivingMethod,
			NetworkName:     transfer.ReceivingNetworkName,
			BankName:        transfer.ReceivingBankName,
			AccountNumber:   transfer.ReceivingBankAccount,
		},
		Notes:     transfer.Notes,
		ExpiresAt: transfer.ExpiresAt,
		CreatedAt: transfer.CreatedAt,
	}

	return resp, nil
}
