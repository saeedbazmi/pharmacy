-- First data source: Darukade, limited to a single category listing page.
-- Prices on the site are toman; the ingest pipeline stores rial.

INSERT INTO pharmacies (slug, name, site_domain, status)
VALUES ('darukade', 'داروکده', 'darukade.com', 'active');

INSERT INTO data_sources (pharmacy_id, kind, config, schedule_interval, is_enabled)
SELECT id,
       'crawler',
       '{
         "fetcher": "darukade",
         "listing_url": "https://darukade.com/products/cosmetic-eye-and-lip-anti-wrinkle-763",
         "max_pages": 1
       }'::jsonb,
       INTERVAL '1 hour',
       TRUE
FROM pharmacies
WHERE slug = 'darukade';
