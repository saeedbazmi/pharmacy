-- name: ListSourceHealth :many
SELECT ds.id,
       ds.pharmacy_id,
       ph.name  AS pharmacy_name,
       ph.slug  AS pharmacy_slug,
       ds.kind,
       ds.schedule_interval::text AS schedule_interval,
       ds.is_enabled,
       ds.last_run_at,
       ds.last_status,
       ds.last_error,
       ds.price_reject_count,
       COALESCE(sr.id, 0)::bigint         AS run_id,
       COALESCE(sr.status, '')::text      AS run_status,
       COALESCE(sr.started_at, TIMESTAMPTZ 'epoch') AS run_started_at,
       sr.finished_at                     AS run_finished_at,
       COALESCE(sr.ok_count, 0)::int      AS run_ok_count,
       COALESCE(sr.fail_count, 0)::int    AS run_fail_count,
       sr.error                           AS run_error
FROM data_sources ds
         JOIN pharmacies ph ON ph.id = ds.pharmacy_id AND ph.deleted_at IS NULL
         LEFT JOIN LATERAL (
    SELECT id, status, started_at, finished_at, ok_count, fail_count, error
    FROM sync_runs r
    WHERE r.source_id = ds.id
    ORDER BY r.started_at DESC
    LIMIT 1
    ) sr ON TRUE
WHERE ds.deleted_at IS NULL
ORDER BY ph.name, ds.id;

-- name: ListSyncSuccessRates :many
SELECT source_id,
       COUNT(*)::int                                              AS runs,
       COUNT(*) FILTER (WHERE status = 'succeeded')::int          AS succeeded,
       COUNT(*) FILTER (WHERE status = 'failed')::int             AS failed
FROM sync_runs
WHERE started_at >= sqlc.arg(since)
  AND (sqlc.arg(source_id)::bigint = 0 OR source_id = sqlc.arg(source_id))
GROUP BY source_id;

-- name: ListSourceSyncRuns :many
SELECT id, status, started_at, finished_at, ok_count, fail_count, error
FROM sync_runs
WHERE source_id = $1
ORDER BY started_at DESC
LIMIT sqlc.arg(page_limit);

-- name: ListStaleSources :many
SELECT ds.id,
       ph.name AS pharmacy_name,
       ph.slug AS pharmacy_slug,
       MIN(o.last_seen_at)::timestamptz AS oldest_seen,
       COUNT(o.id)::int    AS offer_count
FROM data_sources ds
         JOIN pharmacies ph ON ph.id = ds.pharmacy_id AND ph.deleted_at IS NULL
         JOIN offers o ON o.pharmacy_id = ds.pharmacy_id AND o.status = 'active'
WHERE ds.deleted_at IS NULL
GROUP BY ds.id, ph.name, ph.slug
ORDER BY MIN(o.last_seen_at) ASC;

-- name: ListSuspiciousOffers :many
SELECT o.id,
       o.product_id,
       o.pharmacy_id,
       o.price_rial,
       o.proposed_price_rial,
       o.proposed_at,
       o.product_url,
       o.last_seen_at,
       o.in_stock,
       p.slug      AS product_slug,
       p.name_fa   AS product_name,
       ph.name     AS pharmacy_name,
       ph.slug     AS pharmacy_slug,
       COALESCE(si.source_id, 0)::bigint AS source_id,
       COALESCE(ds.price_reject_count, 0)::int AS source_reject_count
FROM offers o
         JOIN products p ON p.id = o.product_id AND p.deleted_at IS NULL
         JOIN pharmacies ph ON ph.id = o.pharmacy_id AND ph.deleted_at IS NULL
         LEFT JOIN source_items si ON si.id = o.source_item_id
         LEFT JOIN data_sources ds ON ds.id = si.source_id AND ds.deleted_at IS NULL
WHERE o.status = 'suspicious'
ORDER BY o.proposed_at DESC NULLS LAST, o.id
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountSuspiciousOffers :one
SELECT COUNT(*)::bigint
FROM offers
WHERE status = 'suspicious';

-- name: GetSuspiciousOffer :one
SELECT o.id,
       o.product_id,
       o.pharmacy_id,
       o.price_rial,
       o.proposed_price_rial,
       o.proposed_at,
       o.status,
       o.in_stock,
       o.source_item_id,
       COALESCE(si.source_id, 0)::bigint AS source_id
FROM offers o
         LEFT JOIN source_items si ON si.id = o.source_item_id
WHERE o.id = $1;

-- name: ApproveSuspiciousOffer :one
UPDATE offers
SET price_rial          = proposed_price_rial,
    status              = 'active',
    proposed_price_rial = NULL,
    proposed_at         = NULL
WHERE id = $1
  AND status = 'suspicious'
  AND proposed_price_rial IS NOT NULL
RETURNING id, product_id, pharmacy_id, price_rial, in_stock;

-- name: RejectSuspiciousOffer :exec
UPDATE offers
SET status              = 'active',
    proposed_price_rial = NULL,
    proposed_at         = NULL
WHERE id = $1
  AND status = 'suspicious';

-- name: IncrementSourcePriceRejects :exec
UPDATE data_sources
SET price_reject_count = price_reject_count + 1
WHERE id = $1
  AND deleted_at IS NULL;
