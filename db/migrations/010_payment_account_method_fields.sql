-- +goose Up
ALTER TABLE payment_accounts
    ALTER COLUMN account_number DROP NOT NULL,
    ADD COLUMN phone_number TEXT,
    ADD COLUMN sort_code TEXT,
    ADD COLUMN iban TEXT;

-- +goose Down
UPDATE payment_accounts
SET account_number = COALESCE(account_number, '')
WHERE account_number IS NULL;

ALTER TABLE payment_accounts
    DROP COLUMN IF EXISTS iban,
    DROP COLUMN IF EXISTS sort_code,
    DROP COLUMN IF EXISTS phone_number,
    ALTER COLUMN account_number SET NOT NULL;
