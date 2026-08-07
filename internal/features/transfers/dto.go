package transfers

import "cnmt/internal/infra/db"

type createTransferRequest struct {
    SourceCountryID      int `json:"source_country_id" validate:"gt=0"`
	DestinationCountryID int `json:"destination_country_id" validate:"gt=0"`

    AmountSent string `json:"amount_sent" validate:"required,numeric"`
    AmountReceived string `json:"amount_received" validate:"required,numeric"`

   	SenderPhone string `json:"sender_phone" validate:"required,min=8,max=20"`

    Recipient *recipientDTO `json:"recipient" validate:"required"`

    Notes *string `json:"notes,omitempty" validate:"omitempty,max=500"`
}

type recipientDTO struct {
	RecipientName  string `json:"recipient_name" validate:"required,min=2,max=100"`
    RecipientPhone *string `json:"recipient_phone,omitempty" validate:"omitempty,required_if=ReceivingMethod MOBILE_MONEY"` // only for mobile money

    ReceivingMethod db.ReceivingMethods `json:"receiving_method" validate:"required,oneof=BANK MOBILE_MONEY"`

    ReceivingNetworkID *int `json:"receiving_network_id,omitempty" validate:"omitempty,required_if=ReceivingMethod MOBILE_MONEY"` // only for mobile money

    BankID *int `json:"bank_id,omitempty" validate:"omitempty,required_if=ReceivingMethod BANK"` // only for bank transfer

    AccountNumber *string `json:"account_number,omitempty" validate:"omitempty,alphanum,required_if=ReceivingMethod BANK"`
}

type createTransferResponse struct {
	TransferID string `json:"transfer_id"`
	Reference string `json:"reference"`
	ExpiresIn int64 `json:"expires_in"`
}