package dashboard

import (
	"time"

	"cnmt/internal/infra/db"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	defaultRecentLimit   = 8
	defaultTopRoutes     = 5
	defaultActivityLimit = 10
	expiringWindow       = 24 * time.Hour
	maxRangeDays         = 366
)

type Period struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type MoneyByCurrency struct {
	Currency string          `json:"currency"`
	Amount   decimal.Decimal `json:"amount"`
}

type Overview struct {
	TotalTransfers      int64             `json:"total_transfers"`
	PendingPayment      int64             `json:"pending_payment"`
	PaymentVerification int64             `json:"payment_verification"`
	Processing          int64             `json:"processing"`
	Completed           int64             `json:"completed"`
	Cancelled           int64             `json:"cancelled"`
	Failed              int64             `json:"failed"`
	TotalTransferVolume []MoneyByCurrency `json:"total_transfer_volume"`
	TotalFees           []MoneyByCurrency `json:"total_fees"`
}

type TransferSummary struct {
	ID                   uuid.UUID           `json:"id"`
	Reference            string              `json:"reference"`
	Status               db.TransferStatus   `json:"status"`
	SourceCountry        string              `json:"source_country"`
	DestinationCountry   string              `json:"destination_country"`
	AmountSent           decimal.Decimal     `json:"amount_sent"`
	CurrencyCode         string              `json:"currency_code"`
	CurrencySymbol       string              `json:"currency_symbol"`
	SenderPhone          string              `json:"sender_phone"`
	ReceivingAccountName string              `json:"receiving_account_name"`
	ReceivingMethod      db.ReceivingMethods `json:"receiving_method"`
	ExpiresAt            *time.Time          `json:"expires_at,omitempty"`
	CreatedAt            time.Time           `json:"created_at"`
}

type ActionRequired struct {
	PaymentVerificationCount int64             `json:"payment_verification_count"`
	ProcessingCount          int64             `json:"processing_count"`
	ExpiringCount            int64             `json:"expiring_count"`
	Transfers                []TransferSummary `json:"transfers"`
}

type VolumePoint struct {
	Date          string          `json:"date"`
	Currency      string          `json:"currency"`
	TransferCount int64           `json:"transfer_count"`
	Volume        decimal.Decimal `json:"volume"`
}

type StatusCount struct {
	Status db.TransferStatus `json:"status"`
	Count  int64             `json:"count"`
}

type TopRoute struct {
	RouteID             uuid.UUID       `json:"route_id"`
	SourceCountry       string          `json:"source_country"`
	SourceISOCode       string          `json:"source_iso_code"`
	SourceFlag          string          `json:"source_flag"`
	SourceCurrency      string          `json:"source_currency"`
	DestinationCountry  string          `json:"destination_country"`
	DestinationISOCode  string          `json:"destination_iso_code"`
	DestinationFlag     string          `json:"destination_flag"`
	DestinationCurrency string          `json:"destination_currency"`
	TransferCount       int64           `json:"transfer_count"`
	TransferVolume      decimal.Decimal `json:"transfer_volume"`
}

type ActivityItem struct {
	Status    db.TransferStatus `json:"status"`
	Reference string            `json:"reference"`
	Actor     string            `json:"actor"`
	Note      *string           `json:"note,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type Response struct {
	Period             Period            `json:"period"`
	Overview           Overview          `json:"overview"`
	ActionRequired     ActionRequired    `json:"action_required"`
	Volume             []VolumePoint     `json:"volume"`
	StatusDistribution []StatusCount     `json:"status_distribution"`
	RecentTransfers    []TransferSummary `json:"recent_transfers"`
	TopRoutes          []TopRoute        `json:"top_routes"`
	RecentActivity     []ActivityItem    `json:"recent_activity"`
}
