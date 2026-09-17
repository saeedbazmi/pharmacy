-- name: GetProductBySlug :one
SELECT p.id,
       p.slug,
       p.name_fa,
       p.name_en,
       p.generic_name,
       p.dosage_form,
       p.strength,
       p.image_url,
       p.description,
       p.status,
       p.updated_at,
       b.name_fa AS brand_name,
       c.name_fa AS category_name
FROM products p
         LEFT JOIN brands b ON b.id = p.brand_id AND b.deleted_at IS NULL
         LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
WHERE p.slug = $1
  AND p.deleted_at IS NULL
  AND p.status = 'published';
