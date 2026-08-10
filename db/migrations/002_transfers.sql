-- +goose Up
CREATE TYPE transfer_status AS ENUM (
    'PENDING_PAYMENT',
    'PAYMENT_RECEIVED',
    'VERIFYING',
    'PROCESSING',
    'COMPLETED',
    'FAILED',
    'CANCELLED'
);
CREATE TABLE transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference TEXT NOT NULL,
    route_id UUID NOT NULL REFERENCES routes (id),
    status transfer_status NOT NULL DEFAULT 'PENDING_PAYMENT',
    sender_phone TEXT NOT NULL,
    receiving_account_name TEXT NOT NULL,
    receiving_mobile_money_number TEXT,
    receiving_method receiving_methods NOT NULL,
    receiving_money_network_id UUID REFERENCES payment_channels (id),
    receiving_bank_id UUID REFERENCES payment_channels (id),
    receiving_bank_account TEXT,
    payment_proof_key TEXT,
    exchange_rate NUMERIC(18, 8) NOT NULL,
    fee NUMERIC(18, 2) NOT NULL,
    amount_sent NUMERIC(18, 2) NOT NULL,
    amount_received NUMERIC(18, 2) NOT NULL,
    notes TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX uq_transfers_reference ON transfers (reference);
CREATE INDEX idx_transfers_route_id ON transfers (route_id);
CREATE INDEX idx_transfers_receiving_money_network_id ON transfers (receiving_money_network_id);
CREATE INDEX idx_transfers_receiving_bank_id ON transfers (receiving_bank_id);
-- Sender history / support lookups.
CREATE INDEX idx_transfers_sender_phone_created_at ON transfers (sender_phone, created_at DESC);
-- Expiry worker: pending transfers past expires_at.
CREATE INDEX idx_transfers_pending_expires_at ON transfers (expires_at)
WHERE status = 'PENDING_PAYMENT'
    AND deleted_at IS NULL;
-- +goose Down
DROP TABLE IF EXISTS transfers;
DROP TYPE IF EXISTS transfer_status;