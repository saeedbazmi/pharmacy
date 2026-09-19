-- GiST trigram index lets relevance ranking use KNN (document <-> query)
-- and stop after LIMIT instead of sorting every matching row.

CREATE INDEX IF NOT EXISTS products_search_gist_idx
    ON products USING GIST (search_document gist_trgm_ops)
    WHERE deleted_at IS NULL AND status = 'published';
