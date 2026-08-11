-- +goose Up
CREATE TABLE transfer_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transfer_id UUID NOT NULL REFERENCES transfers (id),
    status transfer_status NOT NULL,
    actor TEXT NOT NULL,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_transfer_events_transfer_id ON transfer_events (transfer_id);

-- +goose Down
DROP TABLE IF EXISTS transfer_events;
