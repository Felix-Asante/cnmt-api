-- +goose Up
CREATE TABLE payment_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    country_id BIGINT NOT NULL REFERENCES countries (id),
    payment_method receiving_methods NOT NULL,
    name TEXT NOT NULL,
    account_name TEXT NOT NULL,
    account_number TEXT NOT NULL,
    payment_channel_id UUID REFERENCES payment_channels (id),
    currency_code CHAR(3) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- List active accounts by country / method (route payment instructions).
CREATE INDEX idx_payment_accounts_country_method_active
    ON payment_accounts (country_id, payment_method)
    WHERE is_active = TRUE AND deleted_at IS NULL;

CREATE INDEX idx_payment_accounts_payment_channel_id
    ON payment_accounts (payment_channel_id)
    WHERE payment_channel_id IS NOT NULL;

-- Link + immutable payment-instruction snapshot on transfers.
-- Nullable so existing transfer rows remain valid.
ALTER TABLE transfers
    ADD COLUMN payment_account_id UUID REFERENCES payment_accounts (id),
    ADD COLUMN payment_method receiving_methods,
    ADD COLUMN payment_account_name TEXT,
    ADD COLUMN payment_account_number TEXT,
    ADD COLUMN payment_channel_name TEXT,
    ADD COLUMN payment_currency_code CHAR(3);

CREATE INDEX idx_transfers_payment_account_id
    ON transfers (payment_account_id)
    WHERE payment_account_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_transfers_payment_account_id;

ALTER TABLE transfers
    DROP COLUMN IF EXISTS payment_currency_code,
    DROP COLUMN IF EXISTS payment_channel_name,
    DROP COLUMN IF EXISTS payment_account_number,
    DROP COLUMN IF EXISTS payment_account_name,
    DROP COLUMN IF EXISTS payment_method,
    DROP COLUMN IF EXISTS payment_account_id;

DROP TABLE IF EXISTS payment_accounts;
