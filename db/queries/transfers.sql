-- name: GetTransferByReference :one
SELECT t.*,
    src.id AS source_country_id,
    src.name AS source_country_name,
    src.currency_code AS source_currency_code,
    src.currency_symbol AS source_currency_symbol,
    dst.id AS destination_country_id,
    dst.name AS destination_country_name,
    dst.currency_code AS destination_currency_code,
    dst.currency_symbol AS destination_currency_symbol,
    network.name AS receiving_network_name,
    bank.name AS receiving_bank_name
FROM transfers t
    JOIN routes r ON r.id = t.route_id
    JOIN countries src ON src.id = r.source_country_id
    JOIN countries dst ON dst.id = r.destination_country_id
    LEFT JOIN payment_channels network ON network.id = t.receiving_money_network_id
    LEFT JOIN payment_channels bank ON bank.id = t.receiving_bank_id
WHERE t.reference = $1
    AND t.deleted_at IS NULL;
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
-- name: SetPaymentProofKey :exec
UPDATE transfers
SET payment_proof_key = $1,
    status = 'PAYMENT_RECEIVED'
WHERE reference = $2
    AND status = 'PENDING_PAYMENT'
    AND deleted_at IS NULL;