-- Persian search: trigram + unaccent extensions, a stored search document,
-- and the query log used for "missing products" reports.

CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS search_document TEXT NOT NULL DEFAULT '';

-- Approximate Go textfa.Normalize in SQL so existing rows are searchable
-- before the next ingest pass rewrites the document from the Go helper.
UPDATE products p
SET search_document = lower(btrim(concat_ws(' ',
                                            replace(replace(replace(translate(p.name_fa, '۰۱۲۳۴۵۶۷۸۹٠١٢٣٤٥٦٧٨٩', '01234567890123456789'),
                                                                    'ي', 'ی'), 'ك', 'ک'), '‌', ' '),
                                            NULLIF(p.name_en, ''),
                                            NULLIF(p.generic_name, ''),
                                            NULLIF(b.name_fa, ''),
                                            NULLIF(b.name_en, ''))))
FROM brands b
WHERE b.id = p.brand_id
  AND p.search_document = '';

UPDATE products p
SET search_document = lower(btrim(concat_ws(' ',
                                            replace(replace(replace(translate(p.name_fa, '۰۱۲۳۴۵۶۷۸۹٠١٢٣٤٥٦٧٨٩', '01234567890123456789'),
                                                                    'ي', 'ی'), 'ك', 'ک'), '‌', ' '),
                                            NULLIF(p.name_en, ''),
                                            NULLIF(p.generic_name, ''))))
WHERE p.brand_id IS NULL
  AND p.search_document = '';

CREATE INDEX IF NOT EXISTS products_search_trgm_idx
    ON products USING GIN (search_document gin_trgm_ops)
    WHERE deleted_at IS NULL AND status = 'published';

CREATE INDEX IF NOT EXISTS products_name_normalized_prefix_idx
    ON products (name_normalized text_pattern_ops)
    WHERE deleted_at IS NULL AND status = 'published';

CREATE TABLE search_queries (
    id               BIGSERIAL PRIMARY KEY,
    queried_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    query_normalized TEXT        NOT NULL,
    result_count     INTEGER     NOT NULL,
    duration_ms      INTEGER     NOT NULL
);

CREATE INDEX search_queries_time_idx ON search_queries (queried_at DESC);
CREATE INDEX search_queries_zero_idx ON search_queries (queried_at DESC) WHERE result_count = 0;
CREATE INDEX search_queries_freq_idx ON search_queries (query_normalized, queried_at DESC);

-- Canonical product used to verify ye/kaf/latin folding in live search.
INSERT INTO products (slug, name_fa, name_en, generic_name, status, name_normalized, search_document)
SELECT 'acetaminophen-500',
       'استامینوفن ۵۰۰',
       'Acetaminophen 500',
       'acetaminophen',
       'published',
       'استامینوفن 500',
       'استامینوفن 500 acetaminophen 500 acetaminophen'
WHERE NOT EXISTS (SELECT 1 FROM products WHERE slug = 'acetaminophen-500' AND deleted_at IS NULL);
