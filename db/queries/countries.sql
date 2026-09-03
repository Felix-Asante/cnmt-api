-- name: GetCountryByID :one
SELECT *
FROM countries
WHERE id = $1
    AND is_active = TRUE
    AND deleted_at IS NULL;
-- name: GetAdminCountryByID :one
SELECT *
FROM countries
WHERE id = $1
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
    updated_at = now()
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;
-- name: DeleteCountry :one
UPDATE countries
SET deleted_at = now(),
    is_active = FALSE,
    updated_at = now()
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;
-- name: GetPaymentChannelByCountryID :one
SELECT *
FROM payment_channels
WHERE country_id = $1;
-- name: GetPaymentChannelByID :one
SELECT *
FROM payment_channels
WHERE id = $1;
-- name: GetPaymentChannelsByCountryID :many
SELECT *
FROM payment_channels
WHERE country_id = $1
    AND deleted_at IS NULL
ORDER BY channel_type,
    name;
-- name: CreatePaymentChannel :one
INSERT INTO payment_channels (name, channel_type, country_id)
VALUES ($1, $2, $3)
RETURNING *;
-- name: UpdatePaymentChannel :one
UPDATE payment_channels
SET name = $2,
    channel_type = $3,
    updated_at = now()
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;
-- name: DeletePaymentChannel :one
UPDATE payment_channels
SET deleted_at = now(),
    is_active = FALSE,
    updated_at = now()
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;
-- name: GetActiveRouteByCountries :one
SELECT r.*,
    source.name AS source_country_name,
    source.currency_symbol AS source_currency_symbol,
    destination.name AS destination_country_name,
    destination.currency_symbol AS destination_currency_symbol
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
-- name: CreateRoute :one
INSERT INTO routes (
        source_country_id,
        destination_country_id,
        default_exchange_rate,
        fee,
        fee_type,
        min_transfer_amount,
        max_transfer_amount
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
-- name: UpdateRoute :one
UPDATE routes
SET default_exchange_rate = $2,
    fee = $3,
    fee_type = $4,
    min_transfer_amount = $5,
    max_transfer_amount = $6,
    updated_at = now()
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;
-- name: DeleteRoute :one
UPDATE routes
SET deleted_at = now(),
    is_active = FALSE,
    updated_at = now()
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;
-- name: ToggleRouteActive :one
UPDATE routes
SET is_active = NOT is_active,
    updated_at = now()
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;
-- name: ListRoutes :many
SELECT *
FROM routes
WHERE deleted_at IS NULL
    AND source_country_id = COALESCE(NULLIF(sqlc.arg(source_country_id)::bigint, 0), source_country_id)
    AND destination_country_id = COALESCE(NULLIF(sqlc.arg(destination_country_id)::bigint, 0), destination_country_id)
    AND is_active = COALESCE(NULLIF(sqlc.arg(is_active)::text, '')::boolean, is_active)
ORDER BY created_at DESC;