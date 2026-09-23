-- +goose Up
ALTER TABLE transfers
    ADD COLUMN payment_received_at TIMESTAMPTZ,
    ADD COLUMN assurance_sent_at TIMESTAMPTZ;

-- Backfill: treat existing paid-or-later transfers as received at updated_at.
UPDATE transfers
SET payment_received_at = updated_at
WHERE payment_received_at IS NULL
    AND status IN (
        'PAYMENT_RECEIVED',
        'VERIFYING',
        'PROCESSING',
        'COMPLETED'
    )
    AND deleted_at IS NULL;

-- Periodic assurance job: one-shot reminder for in-progress paid transfers.
CREATE INDEX idx_transfers_assurance_pending ON transfers (payment_received_at)
WHERE assurance_sent_at IS NULL
    AND status IN (
        'PAYMENT_RECEIVED',
        'VERIFYING',
        'PROCESSING'
    )
    AND deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_transfers_assurance_pending;

ALTER TABLE transfers
    DROP COLUMN IF EXISTS assurance_sent_at,
    DROP COLUMN IF EXISTS payment_received_at;
