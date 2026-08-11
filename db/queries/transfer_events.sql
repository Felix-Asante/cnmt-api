-- name: TransitionTransferStatus :execrows
UPDATE transfers
SET status = $1::transfer_status,
    updated_at = now()
WHERE id = $2
    AND status = $3::transfer_status
    AND deleted_at IS NULL;

-- name: CreateTransferEvent :one
INSERT INTO transfer_events (transfer_id, status, actor, note)
VALUES ($1, $2, $3, $4)
RETURNING id, transfer_id, status, actor, note, created_at;

-- name: GetTransferEventsByTransferID :many
SELECT id, transfer_id, status, actor, note, created_at
FROM transfer_events
WHERE transfer_id = $1
ORDER BY created_at ASC;
