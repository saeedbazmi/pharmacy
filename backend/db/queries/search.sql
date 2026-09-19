-- name: SearchProducts :many
-- GiST KNN (document <-> query) with LIMIT, no selective WHERE on the
-- document: an unselective %/LIKE clause would force a seq scan of the
-- whole catalogue. The caller keeps rows with sim >= MinSimilarity.
SELECT p.id,
       p.slug,
       p.name_fa,
       p.image_url,
       b.name_fa AS brand_name,
       COALESCE((SELECT MIN(o.price_rial) FROM offers o WHERE o.product_id = p.id AND o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since)), 0)::bigint AS lowest_price_rial,
       (SELECT COUNT(*)::int FROM offers o WHERE o.product_id = p.id AND o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since)) AS offer_count,
       (SELECT COUNT(*)::int FROM offers o WHERE o.product_id = p.id AND o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since) AND o.in_stock) AS in_stock_count,
       similarity(p.search_document, sqlc.arg(q))::float8 AS sim
FROM products p
         LEFT JOIN brands b ON b.id = p.brand_id AND b.deleted_at IS NULL
         LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
WHERE p.deleted_at IS NULL
  AND p.status = 'published'
  AND (sqlc.arg(brand_slug)::text = '' OR b.slug = sqlc.arg(brand_slug))
  AND (sqlc.arg(category_slug)::text = '' OR c.slug = sqlc.arg(category_slug))
  AND (
            NOT sqlc.arg(in_stock_only)::bool
        OR EXISTS (SELECT 1
                   FROM offers ox
                   WHERE ox.product_id = p.id
                     AND ox.status = 'active'
                     AND ox.last_seen_at >= sqlc.arg(fresh_since)
                     AND ox.in_stock)
    )
ORDER BY p.search_document <-> sqlc.arg(q)
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: SearchProductsMatched :many
-- Used when the match set is small enough that a GIN-filtered scan plus
-- sort beats an unfiltered KNN pass (short distinctive queries).
WITH cfg AS (
    SELECT set_limit(sqlc.arg(sim_limit)::real) AS lim
)
SELECT p.id,
       p.slug,
       p.name_fa,
       p.image_url,
       b.name_fa AS brand_name,
       COALESCE((SELECT MIN(o.price_rial) FROM offers o WHERE o.product_id = p.id AND o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since)), 0)::bigint AS lowest_price_rial,
       (SELECT COUNT(*)::int FROM offers o WHERE o.product_id = p.id AND o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since)) AS offer_count,
       (SELECT COUNT(*)::int FROM offers o WHERE o.product_id = p.id AND o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since) AND o.in_stock) AS in_stock_count
FROM cfg,
     products p
         LEFT JOIN brands b ON b.id = p.brand_id AND b.deleted_at IS NULL
         LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
WHERE p.deleted_at IS NULL
  AND p.status = 'published'
  AND (
            p.name_normalized = sqlc.arg(q)
        OR p.name_normalized LIKE sqlc.arg(q) || '%'
        OR p.search_document LIKE '%' || sqlc.arg(q) || '%'
        OR p.search_document OPERATOR (public.%) sqlc.arg(q)
    )
  AND (sqlc.arg(brand_slug)::text = '' OR b.slug = sqlc.arg(brand_slug))
  AND (sqlc.arg(category_slug)::text = '' OR c.slug = sqlc.arg(category_slug))
  AND (
            NOT sqlc.arg(in_stock_only)::bool
        OR EXISTS (SELECT 1
                   FROM offers ox
                   WHERE ox.product_id = p.id
                     AND ox.status = 'active'
                     AND ox.last_seen_at >= sqlc.arg(fresh_since)
                     AND ox.in_stock)
    )
ORDER BY
    CASE
        WHEN p.name_normalized = sqlc.arg(q) THEN 0
        WHEN p.name_normalized LIKE sqlc.arg(q) || '%' THEN 1
        WHEN p.search_document LIKE sqlc.arg(q) || '%' THEN 2
        ELSE 3
        END,
    p.search_document OPERATOR (public.<->) sqlc.arg(q),
    p.id
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: SearchProductsByPrice :many
WITH cfg AS (
    SELECT set_limit(sqlc.arg(sim_limit)::real) AS lim
)
SELECT p.id,
       p.slug,
       p.name_fa,
       p.image_url,
       b.name_fa AS brand_name,
       COALESCE(MIN(o.price_rial) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since)), 0)::bigint AS lowest_price_rial,
       COUNT(o.id) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since))::int AS offer_count,
       COUNT(o.id) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since) AND o.in_stock)::int AS in_stock_count
FROM cfg,
     products p
         LEFT JOIN brands b ON b.id = p.brand_id AND b.deleted_at IS NULL
         LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
         LEFT JOIN offers o ON o.product_id = p.id
WHERE p.deleted_at IS NULL
  AND p.status = 'published'
  AND (
            p.name_normalized = sqlc.arg(q)
        OR p.name_normalized LIKE sqlc.arg(q) || '%'
        OR p.search_document LIKE '%' || sqlc.arg(q) || '%'
        OR p.search_document OPERATOR (public.%) sqlc.arg(q)
    )
  AND (sqlc.arg(brand_slug)::text = '' OR b.slug = sqlc.arg(brand_slug))
  AND (sqlc.arg(category_slug)::text = '' OR c.slug = sqlc.arg(category_slug))
  AND (
            NOT sqlc.arg(in_stock_only)::bool
        OR EXISTS (SELECT 1
                   FROM offers ox
                   WHERE ox.product_id = p.id
                     AND ox.status = 'active'
                     AND ox.last_seen_at >= sqlc.arg(fresh_since)
                     AND ox.in_stock)
    )
GROUP BY p.id, b.name_fa
ORDER BY
    CASE WHEN sqlc.arg(sort)::text = 'price_asc' THEN COALESCE(MIN(o.price_rial) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since)), 0) END ASC NULLS LAST,
    CASE WHEN sqlc.arg(sort)::text = 'price_desc' THEN COALESCE(MIN(o.price_rial) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since)), 0) END DESC NULLS LAST,
    p.id
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountSearchProducts :one
WITH cfg AS (
    SELECT set_limit(sqlc.arg(sim_limit)::real) AS lim
)
SELECT COUNT(*)::bigint AS total_count
FROM cfg,
     products p
         LEFT JOIN brands b ON b.id = p.brand_id AND b.deleted_at IS NULL
         LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
WHERE p.deleted_at IS NULL
  AND p.status = 'published'
  AND (
            p.name_normalized = sqlc.arg(q)
        OR p.name_normalized LIKE sqlc.arg(q) || '%'
        OR p.search_document LIKE '%' || sqlc.arg(q) || '%'
        OR p.search_document OPERATOR (public.%) sqlc.arg(q)
    )
  AND (sqlc.arg(brand_slug)::text = '' OR b.slug = sqlc.arg(brand_slug))
  AND (sqlc.arg(category_slug)::text = '' OR c.slug = sqlc.arg(category_slug))
  AND (
            NOT sqlc.arg(in_stock_only)::bool
        OR EXISTS (SELECT 1
                   FROM offers ox
                   WHERE ox.product_id = p.id
                     AND ox.status = 'active'
                     AND ox.last_seen_at >= sqlc.arg(fresh_since)
                     AND ox.in_stock)
    );

-- name: ListSearchBrands :many
SELECT b.slug, b.name_fa
FROM brands b
WHERE b.deleted_at IS NULL
  AND EXISTS (SELECT 1
              FROM products p
              WHERE p.brand_id = b.id
                AND p.deleted_at IS NULL
                AND p.status = 'published')
ORDER BY b.name_fa
LIMIT 80;

-- name: ListSearchCategories :many
SELECT c.slug, c.name_fa
FROM categories c
WHERE c.deleted_at IS NULL
  AND EXISTS (SELECT 1
              FROM products p
              WHERE p.category_id = c.id
                AND p.deleted_at IS NULL
                AND p.status = 'published')
ORDER BY c.name_fa
LIMIT 80;

-- name: InsertSearchQuery :exec
INSERT INTO search_queries (query_normalized, result_count, duration_ms, user_id)
VALUES ($1, $2, $3, sqlc.narg(user_id));

-- name: ListZeroResultSearches :many
SELECT query_normalized, COUNT(*)::bigint AS hits, MAX(queried_at) AS last_seen
FROM search_queries
WHERE result_count = 0
  AND queried_at >= sqlc.arg(since)
GROUP BY query_normalized
ORDER BY hits DESC, last_seen DESC
LIMIT 50;

-- name: ListFrequentSearches :many
SELECT query_normalized, COUNT(*)::bigint AS hits, MAX(queried_at) AS last_seen
FROM search_queries
WHERE queried_at >= sqlc.arg(since)
GROUP BY query_normalized
ORDER BY hits DESC, last_seen DESC
LIMIT 50;
