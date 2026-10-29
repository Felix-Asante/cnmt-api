-- +goose Up
ALTER TABLE transfers
    ADD COLUMN completion_thanks_sent_at TIMESTAMPTZ;

CREATE INDEX idx_transfers_completion_thanks_pending ON transfers (updated_at)
WHERE completion_thanks_sent_at IS NULL
    AND status = 'COMPLETED'
    AND deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_transfers_completion_thanks_pending;

ALTER TABLE transfers
    DROP COLUMN IF EXISTS completion_thanks_sent_at;
