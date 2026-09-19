-- name: GetActiveOfferByID :one
SELECT o.id,
       o.product_id,
       o.pharmacy_id,
       o.product_url,
       o.status
FROM offers o
WHERE o.id = $1
  AND o.status = 'active';

-- name: InsertRedirectClick :exec
INSERT INTO redirect_clicks (offer_id, product_id, pharmacy_id, referrer)
VALUES ($1, $2, $3, $4);

-- name: EnsureRedirectClicksMonth :exec
SELECT ensure_redirect_clicks_month(sqlc.arg(month_start)::date);
