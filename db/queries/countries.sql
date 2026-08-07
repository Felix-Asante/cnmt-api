-- name: GetCountryByID :one
SELECT *
FROM countries
WHERE id = $1;
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