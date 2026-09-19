-- Second pharmacy: داروخانه روشا, limited to the eye-cream category listing.
-- Prices on the site are toman; the ingest pipeline stores rial.

INSERT INTO pharmacies (slug, name, site_domain, status)
VALUES ('rosha', 'داروخانه روشا', 'roshapharmacy.com', 'active');

INSERT INTO data_sources (pharmacy_id, kind, config, schedule_interval, is_enabled)
SELECT id,
       'crawler',
       '{
         "fetcher": "rosha",
         "listing_url": "https://roshapharmacy.com/skin/skin-care-product/eye-circle-cream",
         "max_pages": 1
       }'::jsonb,
       INTERVAL '1 hour',
       TRUE
FROM pharmacies
WHERE slug = 'rosha';
