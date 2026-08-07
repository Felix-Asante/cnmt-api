-- +goose Up
CREATE TABLE countries (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    iso_code CHAR(2) NOT NULL UNIQUE,
    flag TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    currency_name TEXT NOT NULL,
    currency_code CHAR(3) NOT NULL,
    currency_symbol TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE TYPE fee_type AS ENUM ('fixed', 'percentage');
CREATE TABLE routes(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_country_id BIGSERIAL NOT NULL REFERENCES countries(id),
    destination_country_id BIGSERIAL NOT NULL REFERENCES countries(id),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    default_exchange_rate NUMERIC(10, 4) NOT NULL,
    fee NUMERIC(18, 2) NOT NULL,
    fee_type fee_type NOT NULL DEFAULT 'fixed',
    estimated_minutes INTEGER,
    max_transfer_amount NUMERIC(18, 2) NOT NULL,
    min_transfer_amount NUMERIC(18, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
-- +goose Down
DROP TABLE countries;
DROP TABLE routes;
DROP TYPE fee_type;