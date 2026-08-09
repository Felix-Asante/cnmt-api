-- name: GetCountryByID :one
SELECT *
FROM countries
WHERE id = $1
    AND is_active = TRUE
    AND deleted_at IS NULL;
-- name: ListCountries :many
SELECT *
FROM countries
ORDER BY name;
-- name: CreateCountry :one
INSERT INTO countries (name, iso_code, flag, is_active)
VALUES ($1, $2, $3, $4)
RETURNING *;
-- name: UpdateCountry :one
UPDATE countries
SET name = $2,
    iso_code = $3,
    flag = $4,
    is_active = $5
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