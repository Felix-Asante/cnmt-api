-- name: CreatePromoCode :one
INSERT INTO promo_codes (
        code,
        discount_percentage,
        start_date,
        end_date,
        max_uses,
        max_uses_per_user
    )
VALUES (
        UPPER(trim(sqlc.arg(code))),
        sqlc.arg(discount_percentage),
        sqlc.arg(start_date),
        sqlc.arg(end_date),
        sqlc.arg(max_uses),
        sqlc.arg(max_uses_per_user)
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
-- name: UpdatePromoCode :one
UPDATE promo_codes
SET discount_percentage = $2,
    start_date = $3,
    end_date = $4,
    max_uses = $5,
    max_uses_per_user = $6,
    updated_at = now()
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;
-- name: GetPromoCodeByCodeForUpdate :one
SELECT *
FROM promo_codes
WHERE code = UPPER(trim(sqlc.arg(code)))
    AND deleted_at IS NULL
FOR UPDATE;
-- name: RedeemPromoCode :one
INSERT INTO promo_code_redemptions (
        promo_code_id,
        transfer_id,
        sender_phone,
        discount_percentage
    )
VALUES ($1, $2, $3, $4)
RETURNING id;
-- name: CountRedemptionsByPromoCodeID :one
SELECT COUNT(*)::bigint
FROM promo_code_redemptions
WHERE promo_code_id = $1;
-- name: CountRedemptionsByPromoCodeAndSender :one
SELECT COUNT(*)::bigint
FROM promo_code_redemptions
WHERE promo_code_id = $1
    AND sender_phone = $2;
