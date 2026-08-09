-- name: InsertIdempotencyKey :one
INSERT INTO idempotency_keys (key, actor_id, request_hash, status, expires_at)
VALUES ($1, $2, $3, 'processing', $4)
ON CONFLICT (actor_id, key) DO NOTHING
RETURNING *;

-- name: GetIdempotencyKey :one
SELECT *
FROM idempotency_keys
WHERE actor_id = $1
  AND key = $2
  AND expires_at > now();

-- name: CompleteIdempotencyKey :exec
UPDATE idempotency_keys
SET status = 'completed',
    response_code = $3,
    response_body = $4,
    transfer_id = $5
WHERE actor_id = $1
  AND key = $2;
