-- name: UpsertFavorite :exec
INSERT INTO favorites (user_id, product_id)
VALUES ($1, $2)
ON CONFLICT (user_id, product_id) DO NOTHING;

-- name: DeleteFavorite :exec
DELETE FROM favorites
WHERE user_id = $1
  AND product_id = $2;

-- name: FavoriteExists :one
SELECT EXISTS (
    SELECT 1 FROM favorites WHERE user_id = $1 AND product_id = $2
) AS exists;

-- name: ListFavorites :many
SELECT p.id,
       p.slug,
       p.name_fa,
       p.image_url,
       COALESCE(MIN(o.price_rial) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since)), 0)::bigint AS lowest_price_rial,
       COUNT(o.id) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since))::int AS offer_count,
       f.created_at
FROM favorites f
         JOIN products p ON p.id = f.product_id AND p.deleted_at IS NULL AND p.status = 'published'
         LEFT JOIN offers o ON o.product_id = p.id
WHERE f.user_id = sqlc.arg(user_id)
GROUP BY p.id, f.created_at
ORDER BY f.created_at DESC;

-- name: PublishedProductExists :one
SELECT EXISTS (
    SELECT 1 FROM products WHERE id = $1 AND deleted_at IS NULL AND status = 'published'
) AS exists;

-- name: InsertPriceAlert :one
INSERT INTO price_alerts (user_id, product_id, kind, target_price_rial, baseline_price_rial, stock_was_available)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, product_id, kind, target_price_rial, baseline_price_rial, stock_was_available, status, created_at;

-- name: CountActiveAlerts :one
SELECT COUNT(*)::bigint
FROM price_alerts
WHERE user_id = $1
  AND status = 'active'
  AND deleted_at IS NULL;

-- name: ListUserAlerts :many
SELECT a.id, a.user_id, a.product_id, a.kind, a.target_price_rial, a.baseline_price_rial,
       a.stock_was_available, a.status, a.created_at,
       p.slug AS product_slug, p.name_fa AS product_name
FROM price_alerts a
         JOIN products p ON p.id = a.product_id AND p.deleted_at IS NULL
WHERE a.user_id = $1
  AND a.deleted_at IS NULL
ORDER BY a.created_at DESC;

-- name: SoftDeleteAlert :execrows
UPDATE price_alerts
SET deleted_at = now(),
    status     = 'paused'
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL;

-- name: ListActiveAlertsForEval :many
SELECT a.id, a.user_id, a.product_id, a.kind, a.target_price_rial, a.baseline_price_rial,
       a.stock_was_available, u.phone, p.name_fa AS product_name
FROM price_alerts a
         JOIN users u ON u.id = a.user_id AND u.deleted_at IS NULL
         JOIN products p ON p.id = a.product_id AND p.deleted_at IS NULL
WHERE a.status = 'active'
  AND a.deleted_at IS NULL;

-- name: ListProductQuotes :many
SELECT p.id AS product_id,
       COALESCE(MIN(o.price_rial) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since) AND o.in_stock), 0)::bigint AS lowest_in_stock_rial,
       COALESCE(BOOL_OR(o.in_stock) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since)), FALSE)::bool AS in_stock
FROM products p
         LEFT JOIN offers o ON o.product_id = p.id
WHERE p.id = ANY (sqlc.arg(product_ids)::bigint[])
GROUP BY p.id;

-- name: UpdateAlertBaseline :exec
UPDATE price_alerts
SET baseline_price_rial = $2
WHERE id = $1;

-- name: UpdateAlertStockFlag :exec
UPDATE price_alerts
SET stock_was_available = $2
WHERE id = $1;

-- name: InsertAlertDelivery :execrows
INSERT INTO alert_deliveries (alert_id, event_key, status)
VALUES ($1, $2, 'pending')
ON CONFLICT (alert_id, event_key) DO NOTHING;

-- name: ListRetryableDeliveries :many
SELECT d.id, d.alert_id, d.event_key, d.attempts, a.user_id, a.product_id, a.kind, u.phone, p.name_fa AS product_name
FROM alert_deliveries d
         JOIN price_alerts a ON a.id = d.alert_id AND a.deleted_at IS NULL
         JOIN users u ON u.id = a.user_id AND u.deleted_at IS NULL
         JOIN products p ON p.id = a.product_id
WHERE d.status IN ('pending', 'failed')
  AND d.attempts < 5
ORDER BY d.id
LIMIT 200;

-- name: MarkDeliverySent :exec
UPDATE alert_deliveries
SET status     = 'sent',
    attempts   = attempts + 1,
    sent_at    = now(),
    last_error = NULL
WHERE id = $1;

-- name: MarkDeliveryFailed :exec
UPDATE alert_deliveries
SET status     = 'failed',
    attempts   = attempts + 1,
    last_error = $2
WHERE id = $1;
