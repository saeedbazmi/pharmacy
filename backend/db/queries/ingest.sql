-- name: UpsertSourceItem :one
INSERT INTO source_items (source_id, external_id, product_id, raw, fetched_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (source_id, external_id)
    DO UPDATE SET raw        = EXCLUDED.raw,
                  fetched_at = EXCLUDED.fetched_at,
                  product_id = COALESCE(source_items.product_id, EXCLUDED.product_id)
RETURNING id, product_id;

-- name: LinkSourceItemProduct :exec
UPDATE source_items
SET product_id = $2
WHERE id = $1;

-- name: InsertMatchCandidate :exec
INSERT INTO match_candidates (source_item_id, suggested_product_id, score, status, reason)
VALUES ($1, $2, $3, 'pending', $4)
ON CONFLICT (source_item_id) WHERE status = 'pending'
    DO UPDATE SET suggested_product_id = EXCLUDED.suggested_product_id,
                  score                = EXCLUDED.score,
                  reason               = EXCLUDED.reason;
