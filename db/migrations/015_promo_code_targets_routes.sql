-- +goose Up
-- SEND_AMOUNT:      discount is taken off what the sender pays.
-- RECIPIENT_AMOUNT: sender pays in full; the discount value, converted at the
--                   route rate, is added to what the recipient gets.
CREATE TYPE promo_discount_target AS ENUM ('SEND_AMOUNT', 'RECIPIENT_AMOUNT');

ALTER TABLE promo_codes
    ADD COLUMN applies_to_all_routes BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN discount_target promo_discount_target NOT NULL DEFAULT 'SEND_AMOUNT';

-- A row overrides the code's behaviour on that route. When
-- applies_to_all_routes is FALSE, the code only works on routes listed here.
CREATE TABLE promo_code_routes (
    promo_code_id UUID NOT NULL REFERENCES promo_codes (id) ON DELETE CASCADE,
    route_id UUID NOT NULL REFERENCES routes (id),
    discount_target promo_discount_target NOT NULL,
    -- NULL = use promo_codes.discount_percentage.
    discount_percentage NUMERIC(5, 2) CHECK (
        discount_percentage > 0
        AND discount_percentage <= 100
    ),
    PRIMARY KEY (promo_code_id, route_id)
);

-- What the promo actually did to the transfer.
ALTER TABLE promo_code_redemptions
    ADD COLUMN discount_target promo_discount_target NOT NULL DEFAULT 'SEND_AMOUNT',
    -- Source currency, taken off what the sender paid.
    ADD COLUMN discount_amount NUMERIC(18, 2) NOT NULL DEFAULT 0,
    -- Destination currency, added to what the recipient gets.
    ADD COLUMN bonus_amount NUMERIC(18, 2) NOT NULL DEFAULT 0;

-- Backfill existing redemptions: amount_sent was stored after the discount,
-- so discount = amount_sent * pct / (100 - pct).
UPDATE promo_code_redemptions pcr
SET discount_amount = ROUND(
        t.amount_sent * pcr.discount_percentage / (100 - pcr.discount_percentage),
        2
    )
FROM transfers t
WHERE t.id = pcr.transfer_id
    AND pcr.discount_percentage < 100;

-- +goose Down
ALTER TABLE promo_code_redemptions
    DROP COLUMN bonus_amount,
    DROP COLUMN discount_amount,
    DROP COLUMN discount_target;
DROP TABLE promo_code_routes;
ALTER TABLE promo_codes
    DROP COLUMN discount_target,
    DROP COLUMN applies_to_all_routes;
DROP TYPE promo_discount_target;
