-- +goose Up
ALTER TABLE payment_channels
    ADD COLUMN extra_fee NUMERIC(18, 2);

ALTER TABLE payment_channels
    ADD CONSTRAINT payment_channels_extra_fee_non_negative
    CHECK (extra_fee IS NULL OR extra_fee >= 0);

-- +goose Down
ALTER TABLE payment_channels
    DROP CONSTRAINT IF EXISTS payment_channels_extra_fee_non_negative;

ALTER TABLE payment_channels
    DROP COLUMN IF EXISTS extra_fee;
