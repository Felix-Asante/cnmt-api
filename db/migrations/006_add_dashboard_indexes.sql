-- +goose Up
-- Speeds dashboard date-range scans and recent-transfer ordering.
CREATE INDEX IF NOT EXISTS idx_transfers_created_at_alive ON transfers (created_at DESC)
WHERE deleted_at IS NULL;

-- Speeds status distribution / action-required filters within live transfers.
CREATE INDEX IF NOT EXISTS idx_transfers_status_alive ON transfers (status)
WHERE deleted_at IS NULL;

-- Speeds recent admin activity feed.
CREATE INDEX IF NOT EXISTS idx_transfer_events_created_at ON transfer_events (created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_transfer_events_created_at;
DROP INDEX IF EXISTS idx_transfers_status_alive;
DROP INDEX IF EXISTS idx_transfers_created_at_alive;
