-- +goose Up
CREATE TABLE promo_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL,
    discount_percentage NUMERIC(5, 2) NOT NULL CHECK (
        discount_percentage >= 0
        AND discount_percentage <= 100
    ),
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    max_uses INTEGER NOT NULL CHECK (max_uses > 0),
    max_uses_per_user INTEGER NOT NULL CHECK (
        max_uses_per_user > 0
        AND max_uses_per_user <= max_uses
    ),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CHECK (start_date < end_date),
    CHECK (
        code = UPPER(btrim(code))
        AND code <> ''
    )
);
-- Allows recreating a code after soft-delete
CREATE UNIQUE INDEX uq_promo_codes_code_alive ON promo_codes (code)
WHERE deleted_at IS NULL;
CREATE TABLE promo_code_redemptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    promo_code_id UUID NOT NULL REFERENCES promo_codes (id),
    transfer_id UUID NOT NULL UNIQUE REFERENCES transfers (id),
    sender_phone TEXT NOT NULL,
    discount_percentage NUMERIC(5, 2) NOT NULL CHECK (
        discount_percentage >= 0
        AND discount_percentage <= 100
    ),
    used_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- CountRedemptionsByPromoCodeID / CountRedemptionsByPromoCodeAndSender.
CREATE INDEX idx_promo_code_redemptions_code_sender ON promo_code_redemptions (promo_code_id, sender_phone);
-- +goose Down
DROP TABLE IF EXISTS promo_code_redemptions;
DROP TABLE IF EXISTS promo_codes;