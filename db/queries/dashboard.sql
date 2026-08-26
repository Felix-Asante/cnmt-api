-- name: DashboardStatusCounts :many
SELECT t.status,
    COUNT(*)::bigint AS count
FROM transfers t
WHERE t.deleted_at IS NULL
    AND t.created_at >= sqlc.arg(from_ts)::timestamptz
    AND t.created_at < sqlc.arg(to_ts)::timestamptz
GROUP BY t.status
ORDER BY t.status;

-- name: DashboardMoneyTotals :many
SELECT src.currency_code AS source_currency_code,
    dst.currency_code AS destination_currency_code,
    COUNT(*)::bigint AS transfer_count,
    COALESCE(SUM(t.amount_sent), 0)::numeric AS total_amount_sent,
    COALESCE(SUM(t.fee), 0)::numeric AS total_fees
FROM transfers t
    JOIN routes r ON r.id = t.route_id
    JOIN countries src ON src.id = r.source_country_id
    JOIN countries dst ON dst.id = r.destination_country_id
WHERE t.deleted_at IS NULL
    AND t.created_at >= sqlc.arg(from_ts)::timestamptz
    AND t.created_at < sqlc.arg(to_ts)::timestamptz
GROUP BY src.currency_code,
    dst.currency_code
ORDER BY src.currency_code,
    dst.currency_code;

-- name: DashboardDailyVolume :many
SELECT (DATE_TRUNC('day', t.created_at AT TIME ZONE 'UTC'))::date AS day,
    src.currency_code AS currency_code,
    COUNT(*)::bigint AS transfer_count,
    COALESCE(SUM(t.amount_sent), 0)::numeric AS volume
FROM transfers t
    JOIN routes r ON r.id = t.route_id
    JOIN countries src ON src.id = r.source_country_id
WHERE t.deleted_at IS NULL
    AND t.created_at >= sqlc.arg(from_ts)::timestamptz
    AND t.created_at < sqlc.arg(to_ts)::timestamptz
GROUP BY day,
    src.currency_code
ORDER BY day,
    src.currency_code;

-- name: DashboardActionRequiredCounts :one
SELECT COUNT(*) FILTER (
        WHERE t.status IN ('PAYMENT_RECEIVED', 'VERIFYING')
    )::bigint AS payment_verification_count,
    COUNT(*) FILTER (
        WHERE t.status = 'PROCESSING'
    )::bigint AS processing_count,
    COUNT(*) FILTER (
        WHERE t.status = 'PENDING_PAYMENT'
            AND t.expires_at <= sqlc.arg(expiring_before)::timestamptz
    )::bigint AS expiring_count
FROM transfers t
WHERE t.deleted_at IS NULL;

-- name: DashboardActionRequiredTransfers :many
SELECT t.id,
    t.reference,
    t.status,
    t.amount_sent,
    t.sender_phone,
    t.receiving_account_name,
    t.receiving_method,
    t.expires_at,
    t.created_at,
    src.name AS source_country_name,
    src.currency_code AS source_currency_code,
    src.currency_symbol AS source_currency_symbol,
    dst.name AS destination_country_name,
    dst.currency_code AS destination_currency_code
FROM transfers t
    JOIN routes r ON r.id = t.route_id
    JOIN countries src ON src.id = r.source_country_id
    JOIN countries dst ON dst.id = r.destination_country_id
WHERE t.deleted_at IS NULL
    AND (
        t.status IN (
            'PAYMENT_RECEIVED',
            'VERIFYING',
            'PROCESSING'
        )
        OR (
            t.status = 'PENDING_PAYMENT'
            AND t.expires_at <= sqlc.arg(expiring_before)::timestamptz
        )
    )
ORDER BY t.created_at DESC
LIMIT sqlc.arg(row_limit);

-- name: DashboardRecentTransfers :many
SELECT t.id,
    t.reference,
    t.status,
    t.amount_sent,
    t.sender_phone,
    t.receiving_account_name,
    t.receiving_method,
    t.created_at,
    src.name AS source_country_name,
    src.currency_code AS source_currency_code,
    src.currency_symbol AS source_currency_symbol,
    dst.name AS destination_country_name,
    dst.currency_code AS destination_currency_code
FROM transfers t
    JOIN routes r ON r.id = t.route_id
    JOIN countries src ON src.id = r.source_country_id
    JOIN countries dst ON dst.id = r.destination_country_id
WHERE t.deleted_at IS NULL
    AND t.created_at >= sqlc.arg(from_ts)::timestamptz
    AND t.created_at < sqlc.arg(to_ts)::timestamptz
ORDER BY t.created_at DESC
LIMIT sqlc.arg(row_limit);

-- name: DashboardTopRoutes :many
-- Ranked by transfer_count (most-used corridors in the period).
SELECT r.id AS route_id,
    src.id AS source_country_id,
    src.name AS source_country_name,
    src.iso_code AS source_iso_code,
    src.flag AS source_flag,
    src.currency_code AS source_currency_code,
    dst.id AS destination_country_id,
    dst.name AS destination_country_name,
    dst.iso_code AS destination_iso_code,
    dst.flag AS destination_flag,
    dst.currency_code AS destination_currency_code,
    COUNT(*)::bigint AS transfer_count,
    COALESCE(SUM(t.amount_sent), 0)::numeric AS transfer_volume
FROM transfers t
    JOIN routes r ON r.id = t.route_id
    JOIN countries src ON src.id = r.source_country_id
    JOIN countries dst ON dst.id = r.destination_country_id
WHERE t.deleted_at IS NULL
    AND t.created_at >= sqlc.arg(from_ts)::timestamptz
    AND t.created_at < sqlc.arg(to_ts)::timestamptz
GROUP BY r.id,
    src.id,
    src.name,
    src.iso_code,
    src.flag,
    src.currency_code,
    dst.id,
    dst.name,
    dst.iso_code,
    dst.flag,
    dst.currency_code
ORDER BY transfer_count DESC,
    transfer_volume DESC
LIMIT sqlc.arg(row_limit);

-- name: DashboardRecentActivity :many
SELECT e.id,
    e.status,
    e.actor,
    e.note,
    e.created_at,
    t.reference
FROM transfer_events e
    JOIN transfers t ON t.id = e.transfer_id
WHERE t.deleted_at IS NULL
ORDER BY e.created_at DESC
LIMIT sqlc.arg(row_limit);
