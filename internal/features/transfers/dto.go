package transfers

import (
	"time"

	"cnmt/internal/common"
	"cnmt/internal/infra/db"

	"github.com/google/uuid"
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
	CurrencyCode   string `json:"currency_code"`
	CurrencySymbol string `json:"currency_symbol"`
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

func mapTransferToDTO(transfer db.GetTransferByReferenceRow) (getTransferResponse, error) {
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
				CurrencyCode:   transfer.SourceCurrencyCode,
				CurrencySymbol: transfer.SourceCurrencySymbol,
			},
			DestinationCountry: countryDTO{
				ID:             transfer.DestinationCountryID,
				Name:           transfer.DestinationCountryName,
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
