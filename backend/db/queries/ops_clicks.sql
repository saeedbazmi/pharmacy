-- Click reports filter on clicked_at so Postgres can prune monthly partitions.

-- name: CountClicksInRange :one
SELECT COUNT(*)::bigint
FROM redirect_clicks
WHERE clicked_at >= sqlc.arg(range_start)
  AND clicked_at < sqlc.arg(range_end);

-- name: ListClicksByPharmacy :many
SELECT ph.id,
       ph.slug,
       ph.name,
       COUNT(*)::bigint AS clicks
FROM redirect_clicks rc
         JOIN pharmacies ph ON ph.id = rc.pharmacy_id AND ph.deleted_at IS NULL
WHERE rc.clicked_at >= sqlc.arg(range_start)
  AND rc.clicked_at < sqlc.arg(range_end)
GROUP BY ph.id, ph.slug, ph.name
ORDER BY clicks DESC, ph.name;

-- name: ListTopClickedProducts :many
SELECT p.id,
       p.slug,
       p.name_fa,
       COUNT(*)::bigint AS clicks
FROM redirect_clicks rc
         JOIN products p ON p.id = rc.product_id AND p.deleted_at IS NULL
WHERE rc.clicked_at >= sqlc.arg(range_start)
  AND rc.clicked_at < sqlc.arg(range_end)
GROUP BY p.id, p.slug, p.name_fa
ORDER BY clicks DESC, p.name_fa
LIMIT 20;
