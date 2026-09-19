-- name: ListPendingMatches :many
SELECT mc.id,
       mc.source_item_id,
       mc.suggested_product_id,
       COALESCE(mc.score, 0)::float8 AS score,
       mc.status,
       mc.reason,
       mc.created_at,
       si.source_id,
       si.external_id,
       si.raw,
       si.product_id    AS current_product_id,
       p.slug           AS suggested_slug,
       p.name_fa        AS suggested_name,
       p.image_url      AS suggested_image
FROM match_candidates mc
         JOIN source_items si ON si.id = mc.source_item_id
         LEFT JOIN products p ON p.id = mc.suggested_product_id AND p.deleted_at IS NULL
WHERE mc.status = 'pending'
  AND (sqlc.arg(source_id)::bigint = 0 OR si.source_id = sqlc.arg(source_id))
  AND (sqlc.arg(min_score)::float8 <= 0 OR COALESCE(mc.score, 0)::float8 >= sqlc.arg(min_score))
ORDER BY mc.score DESC NULLS LAST, mc.created_at ASC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountPendingMatches :one
SELECT COUNT(*)::bigint AS n
FROM match_candidates mc
         JOIN source_items si ON si.id = mc.source_item_id
WHERE mc.status = 'pending'
  AND (sqlc.arg(source_id)::bigint = 0 OR si.source_id = sqlc.arg(source_id))
  AND (sqlc.arg(min_score)::float8 <= 0 OR COALESCE(mc.score, 0)::float8 >= sqlc.arg(min_score));

-- name: GetMatchCandidate :one
SELECT mc.id,
       mc.source_item_id,
       mc.suggested_product_id,
       COALESCE(mc.score, 0)::float8 AS score,
       mc.status,
       mc.reason,
       mc.previous_product_id,
       mc.linked_product_id,
       si.source_id,
       si.external_id,
       si.raw,
       si.product_id AS current_product_id,
       ds.pharmacy_id
FROM match_candidates mc
         JOIN source_items si ON si.id = mc.source_item_id
         JOIN data_sources ds ON ds.id = si.source_id
WHERE mc.id = $1;

-- name: DecideMatchCandidate :exec
UPDATE match_candidates
SET status              = sqlc.arg(status),
    decided_by          = sqlc.arg(decided_by),
    decided_at          = now(),
    previous_product_id = sqlc.arg(previous_product_id),
    linked_product_id   = sqlc.arg(linked_product_id),
    suggested_product_id = COALESCE(sqlc.arg(suggested_product_id), suggested_product_id)
WHERE id = sqlc.arg(id);

-- name: ReopenMatchCandidate :exec
UPDATE match_candidates
SET status     = 'pending',
    decided_by = NULL,
    decided_at = NULL
WHERE id = sqlc.arg(id);

-- name: GetOfferBySourceItem :one
SELECT id, product_id, pharmacy_id, price_rial, in_stock, product_url, status
FROM offers
WHERE source_item_id = $1
LIMIT 1;

-- name: RelinkOfferProduct :exec
UPDATE offers
SET product_id = sqlc.arg(product_id)
WHERE id = sqlc.arg(id);

-- name: DeleteOffer :exec
DELETE FROM offers
WHERE id = $1;

-- name: CountMatchPendingKPI :one
SELECT COUNT(*)::bigint AS n
FROM match_candidates
WHERE status = 'pending';

-- name: GetSourceItemByID :one
SELECT id, source_id, external_id, product_id, raw
FROM source_items
WHERE id = $1;
