-- name: GetOfferByProductPharmacy :one
SELECT id, price_rial, in_stock, status
FROM offers
WHERE product_id = $1
  AND pharmacy_id = $2;

-- name: UpsertOffer :one
INSERT INTO offers (product_id, pharmacy_id, source_item_id, price_rial, in_stock, product_url, status, last_seen_at,
                    proposed_price_rial, proposed_at)
VALUES ($1, $2, $3, $4, $5, $6, 'active', now(), NULL, NULL)
ON CONFLICT (product_id, pharmacy_id)
    DO UPDATE SET source_item_id       = EXCLUDED.source_item_id,
                  price_rial           = EXCLUDED.price_rial,
                  in_stock             = EXCLUDED.in_stock,
                  product_url          = EXCLUDED.product_url,
                  last_seen_at         = now(),
                  status               = 'active',
                  proposed_price_rial  = NULL,
                  proposed_at          = NULL
RETURNING id, price_rial, in_stock;

-- name: MarkOfferSuspicious :exec
UPDATE offers
SET status              = 'suspicious',
    proposed_price_rial = $2,
    proposed_at         = now(),
    last_seen_at        = now(),
    in_stock            = $3,
    product_url         = $4,
    source_item_id      = $5
WHERE id = $1;

-- name: InsertPriceHistory :exec
INSERT INTO price_history (offer_id, product_id, pharmacy_id, price_rial, in_stock)
VALUES ($1, $2, $3, $4, $5);

-- name: ListActiveOffersByProduct :many
SELECT o.id,
       o.price_rial,
       o.in_stock,
       o.product_url,
       o.last_seen_at,
       o.status,
       ph.name AS pharmacy_name,
       ph.slug AS pharmacy_slug
FROM offers o
         JOIN pharmacies ph ON ph.id = o.pharmacy_id AND ph.deleted_at IS NULL
WHERE o.product_id = $1
  AND o.status = 'active'
ORDER BY o.price_rial ASC, o.last_seen_at DESC;
