-- name: GetBrandBySlug :one
SELECT id, slug, name_fa, name_en
FROM brands
WHERE slug = $1
  AND deleted_at IS NULL;

-- name: InsertBrand :one
INSERT INTO brands (slug, name_fa, name_en)
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetProductByGTIN :one
SELECT id, slug, name_fa
FROM products
WHERE gtin = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetProductByIRC :one
SELECT id, slug, name_fa
FROM products
WHERE irc = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetProductByNormalizedName :one
SELECT id, slug, name_fa
FROM products
WHERE name_normalized = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: InsertProduct :one
INSERT INTO products (slug, name_fa, name_en, image_url, brand_id, gtin, irc, status, name_normalized, search_document)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'published', $8, $9)
RETURNING id;

-- name: SlugExists :one
SELECT EXISTS (
    SELECT 1 FROM products WHERE slug = $1 AND deleted_at IS NULL
) AS exists;

-- name: GetProductLocks :one
SELECT id, locked_fields, name_fa, name_en, generic_name, image_url, source_snapshot
FROM products
WHERE id = $1
  AND deleted_at IS NULL;

-- name: SetProductSourceSnapshot :exec
UPDATE products
SET source_snapshot = $2
WHERE id = $1;

-- name: UpdateProductFromSource :exec
UPDATE products
SET name_fa = CASE
                  WHEN 'name_fa' = ANY (locked_fields) THEN name_fa
                  ELSE sqlc.arg(name_fa)
              END,
    name_en = CASE
                  WHEN 'name_en' = ANY (locked_fields) THEN name_en
                  ELSE sqlc.arg(name_en)
              END,
    image_url = CASE
                    WHEN 'image_url' = ANY (locked_fields) THEN image_url
                    ELSE sqlc.arg(image_url)
                END,
    source_snapshot = sqlc.arg(source_snapshot),
    name_normalized = CASE
                          WHEN 'name_fa' = ANY (locked_fields) THEN name_normalized
                          ELSE sqlc.arg(name_normalized)
                      END,
    search_document = CASE
                          WHEN 'name_fa' = ANY (locked_fields) THEN search_document
                          ELSE sqlc.arg(search_document)
                      END
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;
