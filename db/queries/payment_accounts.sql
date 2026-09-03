-- name: ListActivePaymentAccountsByCountryID :many
SELECT pa.id,
    pa.country_id,
    pa.payment_method,
    pa.name,
    pa.account_name,
    pa.account_number,
    pa.phone_number,
    pa.sort_code,
    pa.iban,
    pa.currency_code,
    pc.name AS channel_name
FROM payment_accounts pa
    LEFT JOIN payment_channels pc ON pc.id = pa.payment_channel_id
    AND pc.deleted_at IS NULL
WHERE pa.country_id = $1
    AND pa.is_active = TRUE
    AND pa.deleted_at IS NULL
ORDER BY pa.payment_method,
    pa.name;

-- name: CreatePaymentAccount :one
INSERT INTO payment_accounts (
        country_id,
        payment_method,
        name,
        account_name,
        account_number,
        phone_number,
        sort_code,
        iban,
        payment_channel_id,
        currency_code
    )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;
