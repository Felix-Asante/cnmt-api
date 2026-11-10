-- name: CreatePromoCode :one
INSERT INTO promo_codes (
        code,
        discount_percentage,
        start_date,
        end_date,
        max_uses,
        max_uses_per_user,
        applies_to_all_routes,
        discount_target
    )
VALUES (
        UPPER(trim(sqlc.arg(code))),
        sqlc.arg(discount_percentage),
        sqlc.arg(start_date),
        sqlc.arg(end_date),
        sqlc.arg(max_uses),
        sqlc.arg(max_uses_per_user),
        sqlc.arg(applies_to_all_routes),
        sqlc.arg(discount_target)
    )
RETURNING *;
-- name: GetPromoCodeByCode :one
SELECT *
FROM promo_codes
WHERE code = UPPER(trim(sqlc.arg(code)))
    AND deleted_at IS NULL;
-- name: GetAllPromoCodes :many
SELECT *
FROM promo_codes
WHERE deleted_at IS NULL
ORDER BY created_at DESC;
-- name: DeletePromoCode :one
UPDATE promo_codes
SET deleted_at = now(),
    updated_at = now()
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;
-- name: GetPromoCodeByID :one
SELECT *
FROM promo_codes
WHERE id = $1
    AND deleted_at IS NULL;
-- name: GetPromoCodeByIDForUpdate :one
SELECT *
FROM promo_codes
WHERE id = $1
    AND deleted_at IS NULL
FOR UPDATE;
-- name: UpdatePromoCode :one
UPDATE promo_codes
SET discount_percentage = sqlc.arg(discount_percentage),
    start_date = sqlc.arg(start_date),
    end_date = sqlc.arg(end_date),
    max_uses = sqlc.arg(max_uses),
    max_uses_per_user = sqlc.arg(max_uses_per_user),
    applies_to_all_routes = sqlc.arg(applies_to_all_routes),
    discount_target = sqlc.arg(discount_target),
    updated_at = now()
WHERE id = sqlc.arg(id)
    AND deleted_at IS NULL
RETURNING *;
-- name: GetPromoCodeByCodeForUpdate :one
SELECT *
FROM promo_codes
WHERE code = UPPER(trim(sqlc.arg(code)))
    AND deleted_at IS NULL
FOR UPDATE;
-- name: RedeemPromoCode :exec
INSERT INTO promo_code_redemptions (
        promo_code_id,
        transfer_id,
        sender_phone,
        discount_percentage,
        discount_target,
        discount_amount,
        bonus_amount
    )
VALUES ($1, $2, $3, $4, $5, $6, $7);
-- Redemptions on cancelled/failed transfers do not count against limits, so an
-- abandoned (expired) transfer releases its use of the code.
-- name: CountRedemptionsByPromoCodeID :one
SELECT COUNT(*)::bigint
FROM promo_code_redemptions pcr
    JOIN transfers t ON t.id = pcr.transfer_id
WHERE pcr.promo_code_id = $1
    AND t.status NOT IN ('CANCELLED', 'FAILED');
-- name: CountRedemptionsByPromoCodeAndSender :one
SELECT COUNT(*)::bigint
FROM promo_code_redemptions pcr
    JOIN transfers t ON t.id = pcr.transfer_id
WHERE pcr.promo_code_id = $1
    AND pcr.sender_phone = $2
    AND t.status NOT IN ('CANCELLED', 'FAILED');
-- name: GetPromoCodeRouteRule :one
SELECT *
FROM promo_code_routes
WHERE promo_code_id = $1
    AND route_id = $2;
-- name: ListPromoCodeRoutes :many
SELECT pcr.*,
    src.id AS source_country_id,
    src.name AS source_country_name,
    dst.id AS destination_country_id,
    dst.name AS destination_country_name
FROM promo_code_routes pcr
    JOIN routes r ON r.id = pcr.route_id
    JOIN countries src ON src.id = r.source_country_id
    JOIN countries dst ON dst.id = r.destination_country_id
WHERE pcr.promo_code_id = ANY(sqlc.arg(promo_code_ids)::uuid [])
ORDER BY src.name,
    dst.name;
-- name: InsertPromoCodeRoute :exec
INSERT INTO promo_code_routes (
        promo_code_id,
        route_id,
        discount_target,
        discount_percentage
    )
VALUES ($1, $2, $3, $4);
-- name: DeletePromoCodeRoutes :exec
DELETE FROM promo_code_routes
WHERE promo_code_id = $1;
-- name: CountAliveRoutesByIDs :one
SELECT COUNT(*)::bigint
FROM routes
WHERE id = ANY(sqlc.arg(route_ids)::uuid [])
    AND deleted_at IS NULL;
