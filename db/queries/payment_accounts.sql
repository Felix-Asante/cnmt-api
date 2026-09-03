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

-- name: GetPaymentAccountByID :one
SELECT pa.id,
    pa.country_id,
    pa.payment_method,
    pa.name,
    pa.account_name,
    pa.account_number,
    pa.phone_number,
    pa.sort_code,
    pa.iban,
    pa.payment_channel_id,
    pa.currency_code,
    pa.is_active,
    pa.created_at,
    pa.updated_at,
    pc.name AS channel_name
FROM payment_accounts pa
    LEFT JOIN payment_channels pc ON pc.id = pa.payment_channel_id
    AND pc.deleted_at IS NULL
WHERE pa.id = $1
    AND pa.deleted_at IS NULL;

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

-- name: UpdatePaymentAccount :one
UPDATE payment_accounts
SET name = $2,
    account_name = $3,
    account_number = $4,
    phone_number = $5,
    sort_code = $6,
    iban = $7,
    payment_channel_id = $8,
    currency_code = $9,
    updated_at = now()
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: SetPaymentAccountActive :one
UPDATE payment_accounts
SET is_active = $2,
    updated_at = now()
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;

-- name: DeletePaymentAccount :one
UPDATE payment_accounts
SET deleted_at = now(),
    is_active = FALSE,
    updated_at = now()
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;
