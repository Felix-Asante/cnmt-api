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
CREATE TYPE receiving_methods AS ENUM ('BANK', 'MOBILE_MONEY');
CREATE TABLE transfers(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference TEXT NOT NULL UNIQUE,
    route_id uuid NOT NULL REFERENCES routes(id),
    status transfer_status NOT NULL DEFAULT 'PENDING_PAYMENT',
    sender_phone TEXT NOT NULL,
    receiving_account_name TEXT NOT NULL,
    receiving_mobile_money_number TEXT,
    receiving_method receiving_methods NOT NULL,
    receiving_money_network_id UUID REFERENCES payment_channels(id),
    receiving_bank_id UUID REFERENCES payment_channels(id),
    receiving_bank_account TEXT,
    payment_proof_key TEXT,
    exchange_rate NUMERIC(18, 8) NOT NULL,
    fee NUMERIC(18, 2) NOT NULL,
    amount_sent NUMERIC(18, 2) NOT NULL,
    amount_received NUMERIC(18, 2) NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ NULL DEFAULT NULL
);
-- +goose Down
DROP TABLE transfers;