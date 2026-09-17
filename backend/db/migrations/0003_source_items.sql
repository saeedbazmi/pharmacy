-- Raw payload of every item as the source returned it. Rows are never deleted:
-- they are the basis for reprocessing and for debugging a broken crawler.

CREATE TABLE source_items (
    id          BIGSERIAL PRIMARY KEY,
    source_id   BIGINT      NOT NULL REFERENCES data_sources (id) ON DELETE CASCADE,
    -- Identifier of the item in the source system.
    external_id TEXT        NOT NULL,
    -- Set once the item is matched to a unified product; NULL while unmatched.
    product_id  BIGINT      REFERENCES products (id) ON DELETE SET NULL,
    raw         JSONB       NOT NULL,
    fetched_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX source_items_external_key ON source_items (source_id, external_id);
CREATE INDEX source_items_product_idx ON source_items (product_id);
CREATE INDEX source_items_unmatched_idx ON source_items (source_id) WHERE product_id IS NULL;
CREATE INDEX source_items_fetched_idx ON source_items (source_id, fetched_at DESC);

CREATE TRIGGER source_items_set_updated_at
    BEFORE UPDATE ON source_items
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
