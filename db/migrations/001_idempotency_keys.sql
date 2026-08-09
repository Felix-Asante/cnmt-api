-- +goose Up
CREATE TABLE idempotency_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key TEXT NOT NULL,
    actor_id TEXT NOT NULL,
    request_hash TEXT NOT NULL,
    status TEXT NOT NULL,
    response_code INT,
    response_body JSONB,
    transfer_id UUID REFERENCES transfers(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    UNIQUE (actor_id, key)
);
CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys (expires_at);
-- +goose Down
DROP TABLE IF EXISTS idempotency_keys;