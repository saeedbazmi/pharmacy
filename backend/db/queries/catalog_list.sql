-- name: ListPublishedProducts :many
SELECT p.id,
       p.slug,
       p.name_fa,
       p.image_url,
       b.name_fa AS brand_name,
       COALESCE(MIN(o.price_rial) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since)), 0)::bigint AS lowest_price_rial,
       COUNT(o.id) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since))::int AS offer_count,
       COUNT(o.id) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since) AND o.in_stock)::int AS in_stock_count
FROM products p
         LEFT JOIN brands b ON b.id = p.brand_id AND b.deleted_at IS NULL
         LEFT JOIN offers o ON o.product_id = p.id
WHERE p.deleted_at IS NULL
  AND p.status = 'published'
GROUP BY p.id, b.name_fa
ORDER BY p.id
LIMIT sqlc.arg(page_limit);

-- name: GetPublishedProductsBySlugs :many
SELECT p.id,
       p.slug,
       p.name_fa,
       p.image_url,
       b.name_fa AS brand_name,
       COALESCE(MIN(o.price_rial) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since)), 0)::bigint AS lowest_price_rial,
       COUNT(o.id) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since))::int AS offer_count,
       COUNT(o.id) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since) AND o.in_stock)::int AS in_stock_count
FROM products p
         LEFT JOIN brands b ON b.id = p.brand_id AND b.deleted_at IS NULL
         LEFT JOIN offers o ON o.product_id = p.id
WHERE p.deleted_at IS NULL
  AND p.status = 'published'
  AND p.slug = ANY (sqlc.arg(slugs)::text[])
GROUP BY p.id, b.name_fa
ORDER BY p.id;

-- name: GetPublishedCategoryBySlug :one
SELECT id, slug, name_fa, updated_at
FROM categories
WHERE slug = $1
  AND deleted_at IS NULL;

-- name: ListPublishedProductsByCategorySlug :many
SELECT p.id,
       p.slug,
       p.name_fa,
       p.image_url,
       b.name_fa AS brand_name,
       COALESCE(MIN(o.price_rial) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since)), 0)::bigint AS lowest_price_rial,
       COUNT(o.id) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since))::int AS offer_count,
       COUNT(o.id) FILTER (WHERE o.status = 'active' AND o.last_seen_at >= sqlc.arg(fresh_since) AND o.in_stock)::int AS in_stock_count
FROM products p
         JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
         LEFT JOIN brands b ON b.id = p.brand_id AND b.deleted_at IS NULL
         LEFT JOIN offers o ON o.product_id = p.id
WHERE p.deleted_at IS NULL
  AND p.status = 'published'
  AND c.slug = sqlc.arg(category_slug)
GROUP BY p.id, b.name_fa
ORDER BY p.id
LIMIT sqlc.arg(page_limit);

-- name: ListPublishedCategorySlugs :many
SELECT c.slug, c.updated_at
FROM categories c
WHERE c.deleted_at IS NULL
  AND EXISTS (SELECT 1
              FROM products p
              WHERE p.category_id = c.id
                AND p.deleted_at IS NULL
                AND p.status = 'published')
ORDER BY c.name_fa
LIMIT 5000;
