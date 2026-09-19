-- name: GetOpsProduct :one
SELECT p.id,
       p.slug,
       p.name_fa,
       p.name_en,
       p.generic_name,
       p.gtin,
       p.irc,
       p.brand_id,
       p.category_id,
       p.image_url,
       p.status,
       p.locked_fields,
       p.source_snapshot,
       b.slug    AS brand_slug,
       b.name_fa AS brand_name,
       c.slug    AS category_slug,
       c.name_fa AS category_name
FROM products p
         LEFT JOIN brands b ON b.id = p.brand_id AND b.deleted_at IS NULL
         LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
WHERE p.id = $1
  AND p.deleted_at IS NULL;

-- name: ListOpsProducts :many
SELECT p.id,
       p.slug,
       p.name_fa,
       p.status,
       p.image_url,
       p.locked_fields,
       b.name_fa AS brand_name,
       c.name_fa AS category_name
FROM products p
         LEFT JOIN brands b ON b.id = p.brand_id AND b.deleted_at IS NULL
         LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
WHERE p.deleted_at IS NULL
  AND (sqlc.narg(status)::text IS NULL OR p.status = sqlc.narg(status))
  AND (sqlc.narg(q)::text IS NULL OR p.name_fa ILIKE '%' || sqlc.narg(q) || '%')
ORDER BY p.updated_at DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountOpsProducts :one
SELECT COUNT(*)::bigint AS n
FROM products
WHERE deleted_at IS NULL
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status))
  AND (sqlc.narg(q)::text IS NULL OR name_fa ILIKE '%' || sqlc.narg(q) || '%');

-- name: UpdateOpsProduct :exec
UPDATE products
SET name_fa          = sqlc.arg(name_fa),
    name_en          = sqlc.arg(name_en),
    generic_name     = sqlc.arg(generic_name),
    brand_id         = sqlc.narg(brand_id),
    category_id      = sqlc.narg(category_id),
    image_url        = sqlc.arg(image_url),
    status           = sqlc.arg(status),
    locked_fields    = sqlc.arg(locked_fields),
    source_snapshot  = sqlc.arg(source_snapshot),
    name_normalized  = sqlc.arg(name_normalized),
    search_document  = sqlc.arg(search_document)
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: ListOpsCategories :many
SELECT id, parent_id, slug, name_fa, position
FROM categories
WHERE deleted_at IS NULL
ORDER BY position, name_fa;

-- name: GetOpsCategory :one
SELECT id, parent_id, slug, name_fa, position
FROM categories
WHERE id = $1
  AND deleted_at IS NULL;

-- name: GetOpsCategoryBySlug :one
SELECT id, parent_id, slug, name_fa, position
FROM categories
WHERE slug = $1
  AND deleted_at IS NULL;

-- name: InsertOpsCategory :one
INSERT INTO categories (parent_id, slug, name_fa, position)
VALUES (sqlc.narg(parent_id), sqlc.arg(slug), sqlc.arg(name_fa), sqlc.arg(position))
RETURNING id;

-- name: UpdateOpsCategory :exec
UPDATE categories
SET parent_id = sqlc.narg(parent_id),
    slug      = sqlc.arg(slug),
    name_fa   = sqlc.arg(name_fa),
    position  = sqlc.arg(position)
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: SoftDeleteCategory :exec
UPDATE categories
SET deleted_at = now()
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: CountCategoryChildren :one
SELECT COUNT(*)::bigint AS n
FROM categories
WHERE parent_id = $1
  AND deleted_at IS NULL;

-- name: CountCategoryProducts :one
SELECT COUNT(*)::bigint AS n
FROM products
WHERE category_id = $1
  AND deleted_at IS NULL;

-- name: ReassignCategoryProducts :exec
UPDATE products
SET category_id = sqlc.arg(new_id)
WHERE category_id = sqlc.arg(old_id)
  AND deleted_at IS NULL;

-- name: InsertCategorySlugRedirect :exec
INSERT INTO category_slug_redirects (old_slug, category_id)
VALUES ($1, $2)
ON CONFLICT (old_slug) DO UPDATE SET category_id = EXCLUDED.category_id;

-- name: GetCategoryRedirect :one
SELECT old_slug, category_id
FROM category_slug_redirects
WHERE old_slug = $1;

-- name: ListOpsBrands :many
SELECT id, slug, name_fa, name_en
FROM brands
WHERE deleted_at IS NULL
ORDER BY name_fa;

-- name: GetOpsBrand :one
SELECT id, slug, name_fa, name_en
FROM brands
WHERE id = $1
  AND deleted_at IS NULL;

-- name: InsertOpsBrand :one
INSERT INTO brands (slug, name_fa, name_en)
VALUES ($1, $2, $3)
RETURNING id;

-- name: UpdateOpsBrand :exec
UPDATE brands
SET slug    = sqlc.arg(slug),
    name_fa = sqlc.arg(name_fa),
    name_en = sqlc.arg(name_en)
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: SoftDeleteBrand :exec
UPDATE brands
SET deleted_at = now()
WHERE id = sqlc.arg(id)
  AND deleted_at IS NULL;

-- name: ReassignBrandProducts :exec
UPDATE products
SET brand_id = sqlc.arg(new_id)
WHERE brand_id = sqlc.arg(old_id)
  AND deleted_at IS NULL;

-- name: CountBrandProducts :one
SELECT COUNT(*)::bigint AS n
FROM products
WHERE brand_id = $1
  AND deleted_at IS NULL;

-- name: ListPublishedSlugs :many
SELECT slug, updated_at
FROM products
WHERE status = 'published'
  AND deleted_at IS NULL
ORDER BY updated_at DESC
LIMIT 50000;
