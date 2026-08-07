-- name: GetTransferByReference :one
SELECT *
FROM transfers
WHERE reference = $1;