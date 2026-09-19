-- Dashboard, internal users, site settings and feature flags (super_admin).

-- name: DashboardCoverage :one
SELECT
    (SELECT COUNT(*)::bigint
     FROM pharmacies
     WHERE deleted_at IS NULL
       AND status = 'active') AS active_pharmacies,
    (SELECT COUNT(*)::bigint
     FROM products
     WHERE deleted_at IS NULL
       AND status = 'published') AS published_products,
    (SELECT COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (now() - last_seen_at))), 0)::float8
     FROM offers
     WHERE status = 'active') AS median_freshness_seconds;

-- name: CountSearchesInRange :one
SELECT COUNT(*)::bigint
FROM search_queries
WHERE queried_at >= sqlc.arg(range_start)
  AND queried_at < sqlc.arg(range_end);

-- name: ListSearchVolumeByDay :many
SELECT date_trunc('day', queried_at)::timestamptz AS day,
       COUNT(*)::bigint                           AS searches
FROM search_queries
WHERE queried_at >= sqlc.arg(range_start)
  AND queried_at < sqlc.arg(range_end)
GROUP BY 1
ORDER BY 1;

-- name: ListClickVolumeByDay :many
SELECT date_trunc('day', clicked_at)::timestamptz AS day,
       COUNT(*)::bigint                           AS clicks
FROM redirect_clicks
WHERE clicked_at >= sqlc.arg(range_start)
  AND clicked_at < sqlc.arg(range_end)
GROUP BY 1
ORDER BY 1;

-- name: ListFrequentSearchesInRange :many
SELECT query_normalized, COUNT(*)::bigint AS hits, MAX(queried_at) AS last_seen
FROM search_queries
WHERE queried_at >= sqlc.arg(range_start)
  AND queried_at < sqlc.arg(range_end)
GROUP BY query_normalized
ORDER BY hits DESC, last_seen DESC
LIMIT 50;

-- name: ListZeroResultSearchesInRange :many
SELECT query_normalized, COUNT(*)::bigint AS hits, MAX(queried_at) AS last_seen
FROM search_queries
WHERE result_count = 0
  AND queried_at >= sqlc.arg(range_start)
  AND queried_at < sqlc.arg(range_end)
GROUP BY query_normalized
ORDER BY hits DESC, last_seen DESC
LIMIT 50;

-- name: ListInternalUsers :many
SELECT id, username, role, is_active, created_at
FROM internal_users
WHERE deleted_at IS NULL
ORDER BY id;

-- name: CountActiveSuperAdmins :one
SELECT COUNT(*)::bigint
FROM internal_users
WHERE deleted_at IS NULL
  AND is_active
  AND role = 'super_admin';

-- name: UpdateInternalUser :exec
UPDATE internal_users
SET role          = $2,
    is_active     = $3,
    password_hash = COALESCE(sqlc.narg(password_hash), password_hash)
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ListSiteSettings :many
SELECT key, value, updated_at
FROM site_settings
ORDER BY key;

-- name: UpsertSiteSetting :exec
INSERT INTO site_settings (key, value, updated_at)
VALUES ($1, $2, now())
ON CONFLICT (key) DO UPDATE
    SET value      = EXCLUDED.value,
        updated_at = now();

-- name: ListFeatureFlags :many
SELECT id, key, pharmacy_id, enabled, updated_at
FROM feature_flags
ORDER BY key, pharmacy_id NULLS FIRST;

-- name: UpsertGlobalFlag :exec
INSERT INTO feature_flags (key, pharmacy_id, enabled, updated_at)
VALUES ($1, NULL, $2, now())
ON CONFLICT (key) WHERE pharmacy_id IS NULL
    DO UPDATE SET enabled    = EXCLUDED.enabled,
                  updated_at = now();

-- name: UpsertPharmacyFlag :exec
INSERT INTO feature_flags (key, pharmacy_id, enabled, updated_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (key, pharmacy_id) WHERE pharmacy_id IS NOT NULL
    DO UPDATE SET enabled    = EXCLUDED.enabled,
                  updated_at = now();

-- name: ListActivePharmaciesBrief :many
SELECT id, slug, name
FROM pharmacies
WHERE deleted_at IS NULL
  AND status = 'active'
ORDER BY name;
