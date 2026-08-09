package transfers

import (
	"cnmt/internal/infra/db"

	"github.com/google/uuid"
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
	RecipientName  string `json:"recipient_name" validate:"required,min=2,max=100"`
    RecipientPhone *string `json:"recipient_phone,omitempty" validate:"omitempty,e164,required_if=ReceivingMethod MOBILE_MONEY"` // only for mobile money

    ReceivingMethod db.ReceivingMethods `json:"receiving_method" validate:"required,oneof=BANK MOBILE_MONEY"`

    ReceivingNetworkID *uuid.UUID `json:"receiving_network_id,omitempty" validate:"omitempty,uuid,required_if=ReceivingMethod MOBILE_MONEY"` // only for mobile money

    BankID *uuid.UUID `json:"bank_id,omitempty" validate:"omitempty,uuid,required_if=ReceivingMethod BANK"` // only for bank transfer

    AccountNumber *string `json:"account_number,omitempty" validate:"omitempty,alphanum,required_if=ReceivingMethod BANK"`
}

type createTransferResponse struct {
	TransferID uuid.UUID `json:"transfer_id"`
	Reference string `json:"reference"`
	ExpiresIn int64 `json:"expires_in"`
}