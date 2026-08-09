-- name: GetTransferByReference :one
SELECT *
FROM transfers
WHERE reference = $1;
-- name: CreateTransfer :one
INSERT INTO transfers (
        reference,
        route_id,
        status,
        sender_phone,
        receiving_account_name,
        receiving_mobile_money_number,
        receiving_method,
        receiving_money_network_id,
        receiving_bank_id,
        receiving_bank_account,
        payment_proof_key,
        exchange_rate,
        fee,
        amount_sent,
        amount_received,
        notes,
        expires_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15,
        $16,
        $17
    )
RETURNING id;