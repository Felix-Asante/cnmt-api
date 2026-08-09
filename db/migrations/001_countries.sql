-- +goose Up
CREATE TYPE receiving_methods AS ENUM ('BANK', 'MOBILE_MONEY');
CREATE TYPE fee_type AS ENUM ('fixed', 'percentage');

CREATE TABLE countries (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    iso_code CHAR(2) NOT NULL,
    flag TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    currency_name TEXT NOT NULL,
    currency_code CHAR(3) NOT NULL,
    currency_symbol TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- UNIQUE constraints already create indexes; no extra index on PRIMARY KEY (id).
CREATE UNIQUE INDEX uq_countries_name ON countries (name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX uq_countries_iso_code ON countries (iso_code) WHERE deleted_at IS NULL;

CREATE TABLE routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_country_id BIGINT NOT NULL REFERENCES countries (id),
    destination_country_id BIGINT NOT NULL REFERENCES countries (id),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    default_exchange_rate NUMERIC(10, 4) NOT NULL,
    fee NUMERIC(18, 2) NOT NULL,
    fee_type fee_type NOT NULL DEFAULT 'fixed',
    estimated_minutes INTEGER,
    max_transfer_amount NUMERIC(18, 2),
    min_transfer_amount NUMERIC(18, 2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CHECK (source_country_id <> destination_country_id)
);

-- One live corridor per country pair (allows recreating after soft-delete).
CREATE UNIQUE INDEX uq_routes_source_dest_alive
    ON routes (source_country_id, destination_country_id)
    WHERE deleted_at IS NULL;

-- Speeds GetActiveRouteByCountries.
CREATE INDEX idx_routes_active_lookup
    ON routes (source_country_id, destination_country_id)
    WHERE is_active = TRUE AND deleted_at IS NULL;

CREATE TABLE payment_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    channel_type receiving_methods NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    country_id BIGINT NOT NULL REFERENCES countries (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Same network/bank name can exist in different countries; not globally unique.
CREATE UNIQUE INDEX uq_payment_channels_country_name_alive
    ON payment_channels (country_id, name)
    WHERE deleted_at IS NULL;

-- List / filter channels by country + type (many channels per country).
CREATE INDEX idx_payment_channels_country_type_active
    ON payment_channels (country_id, channel_type)
    WHERE is_active = TRUE AND deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS payment_channels;
DROP TABLE IF EXISTS routes;
DROP TABLE IF EXISTS countries;
DROP TYPE IF EXISTS fee_type;
DROP TYPE IF EXISTS receiving_methods;
