-- name: GetCountryByID :one
SELECT *
FROM countries
WHERE id = $1
    AND is_active = TRUE
    AND deleted_at IS NULL;
-- name: GetAllCountries :many
SELECT *
FROM countries
WHERE deleted_at IS NULL
ORDER BY name;
-- name: CreateCountry :one
INSERT INTO countries (
        name,
        iso_code,
        flag,
        currency_name,
        currency_code,
        currency_symbol,
        is_active
    )
VALUES ($1, $2, $3, $4, $5, $6, TRUE)
RETURNING *;
-- name: UpdateCountry :one
UPDATE countries
SET name = $2,
    iso_code = $3,
    flag = $4,
    currency_name = $5,
    currency_code = $6,
    currency_symbol = $7,
    is_active = $8
WHERE id = $1
RETURNING *;
-- name: DeleteCountry :exec
DELETE FROM countries
WHERE id = $1;
-- name: GetPaymentChannelByCountryID :one
SELECT *
FROM payment_channels
WHERE country_id = $1;
-- name: GetPaymentChannelByID :one
SELECT *
FROM payment_channels
WHERE id = $1;
-- name: CreatePaymentChannel :one
INSERT INTO payment_channels (name, channel_type, country_id)
VALUES ($1, $2, $3)
RETURNING *;
-- name: GetActiveRouteByCountries :one
SELECT r.*
FROM routes r
    JOIN countries source ON source.id = r.source_country_id
    JOIN countries destination ON destination.id = r.destination_country_id
WHERE r.source_country_id = $1
    AND r.destination_country_id = $2
    AND r.is_active = TRUE
    AND r.deleted_at IS NULL
    AND source.is_active = TRUE
    AND source.deleted_at IS NULL
    AND destination.is_active = TRUE
    AND destination.deleted_at IS NULL;
-- name: GetActivePCByCountryTypeAndID :one
SELECT pc.*
FROM payment_channels pc
    JOIN countries c ON c.id = pc.country_id
WHERE pc.country_id = $1
    AND pc.channel_type = $2
    AND pc.id = $3
    AND pc.is_active = TRUE
    AND pc.deleted_at IS NULL
    AND c.is_active = TRUE
    AND c.deleted_at IS NULL;
-- name: GetDestCountriesBySrcCountryID :many
SELECT c.id,
    c.name,
    c.iso_code,
    c.flag,
    c.currency_name,
    c.currency_code,
    c.currency_symbol,
    r.min_transfer_amount,
    r.max_transfer_amount,
    r.default_exchange_rate,
    r.fee_type,
    r.fee
FROM countries c
    JOIN routes r ON r.destination_country_id = c.id
    JOIN countries src ON src.id = r.source_country_id
WHERE r.source_country_id = $1
    AND r.is_active = TRUE
    AND r.deleted_at IS NULL
    AND c.is_active = TRUE
    AND c.deleted_at IS NULL
    AND src.is_active = TRUE
    AND src.deleted_at IS NULL
ORDER BY c.name;
-- name: GetAllActiveRouteDestinations :many
SELECT r.source_country_id,
    c.id,
    c.name,
    c.iso_code,
    c.flag,
    c.currency_name,
    c.currency_code,
    c.currency_symbol,
    r.min_transfer_amount,
    r.max_transfer_amount,
    r.default_exchange_rate,
    r.fee_type,
    r.fee
FROM countries c
    JOIN routes r ON r.destination_country_id = c.id
    JOIN countries src ON src.id = r.source_country_id
WHERE r.is_active = TRUE
    AND r.deleted_at IS NULL
    AND c.is_active = TRUE
    AND c.deleted_at IS NULL
    AND src.is_active = TRUE
    AND src.deleted_at IS NULL
ORDER BY r.source_country_id,
    c.name;
-- name: GetActivePaymentChannelsByCountryIDs :many
SELECT id,
    name,
    channel_type,
    country_id
FROM payment_channels
WHERE country_id = ANY($1::bigint [])
    AND is_active = TRUE
    AND deleted_at IS NULL
ORDER BY channel_type,
    name;
-- name: GetAllSourceCountries :many
SELECT DISTINCT c.id,
    c.name,
    c.iso_code,
    c.flag,
    c.currency_name,
    c.currency_code,
    c.currency_symbol
FROM countries c
    JOIN routes r ON r.source_country_id = c.id
WHERE c.is_active = TRUE
    AND c.deleted_at IS NULL
    AND r.is_active = TRUE
    AND r.deleted_at IS NULL
ORDER BY c.name;
-- name: GetCountryByName :one
SELECT *
FROM countries
WHERE name = $1
    AND deleted_at IS NULL;
-- name: DoesPaymentChannelExist :one
SELECT id
FROM payment_channels
WHERE country_id = $1
    AND name = $2
    AND channel_type = $3
    AND is_active = TRUE
    AND deleted_at IS NULL;