-- One offer per (product, pharmacy): the price and stock of a unified product at
-- a single pharmacy. This is the table the comparison table reads from.

CREATE TABLE offers (
    id             BIGSERIAL PRIMARY KEY,
    product_id     BIGINT      NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    pharmacy_id    BIGINT      NOT NULL REFERENCES pharmacies (id) ON DELETE CASCADE,
    source_item_id BIGINT      REFERENCES source_items (id) ON DELETE SET NULL,
    price_rial     BIGINT      NOT NULL CHECK (price_rial >= 0),
    in_stock       BOOLEAN     NOT NULL DEFAULT TRUE,
    -- Deep link to the item on the pharmacy site; the only URL /go/{id} may redirect to.
    product_url    TEXT        NOT NULL,
    -- 'suspicious' offers are withheld from the public site until a human reviews them.
    status         TEXT        NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'suspicious', 'hidden')),
    last_seen_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX offers_product_pharmacy_key ON offers (product_id, pharmacy_id);
-- Drives the price table: active offers of a product ordered by price.
CREATE INDEX offers_product_price_idx ON offers (product_id, price_rial) WHERE status = 'active';
CREATE INDEX offers_pharmacy_idx ON offers (pharmacy_id);
CREATE INDEX offers_stale_idx ON offers (last_seen_at);

CREATE TRIGGER offers_set_updated_at
    BEFORE UPDATE ON offers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Append-only price log: one row per observed change, never per sync run.
CREATE TABLE price_history (
    id          BIGSERIAL PRIMARY KEY,
    offer_id    BIGINT      NOT NULL REFERENCES offers (id) ON DELETE CASCADE,
    product_id  BIGINT      NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    pharmacy_id BIGINT      NOT NULL REFERENCES pharmacies (id) ON DELETE CASCADE,
    price_rial  BIGINT      NOT NULL CHECK (price_rial >= 0),
    in_stock    BOOLEAN     NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX price_history_offer_idx ON price_history (offer_id, recorded_at DESC);
CREATE INDEX price_history_product_idx ON price_history (product_id, recorded_at DESC);
