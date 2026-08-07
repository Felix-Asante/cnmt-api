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
CREATE TYPE receiving_method AS ENUM ('BANK', 'MOBILE_MONEY');
CREATE TABLE transfers(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference TEXT NOT NULL UNIQUE,
    route_id UUID NOT NULL REFERENCES routes(id),
    status transfer_status NOT NULL DEFAULT 'PENDING_PAYMENT',
    sender_name TEXT NOT NULL,
    sender_phone TEXT NOT NULL,
    recipient_name TEXT NOT NULL,
    recipient_phone TEXT NOT NULL,
    receiving_method receiving_method NOT NULL,
    network_name TEXT NOT NULL,
    bank_name TEXT,
    recipient_account TEXT NOT NULL,
    send_amount NUMERIC(18, 2) NOT NULL,
    receive_amount NUMERIC(18, 2) NOT NULL,
    source_currency TEXT NOT NULL,
    destination_currency TEXT NOT NULL,
    exchange_rate NUMERIC(18, 8) NOT NULL,
    fee NUMERIC(18, 2) NOT NULL,
    payment_channel TEXT NOT NULL,
    payment_proof_key TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ NULL DEFAULT NULL
);
-- +goose Down
DROP TABLE transfers;