-- Unified catalog: one product row per real-world item, regardless of how many
-- pharmacies sell it. Offers (0004) hang off these rows.

CREATE TABLE brands (
    id         BIGSERIAL PRIMARY KEY,
    slug       TEXT        NOT NULL,
    name_fa    TEXT        NOT NULL,
    name_en    TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX brands_slug_key ON brands (slug) WHERE deleted_at IS NULL;

CREATE TRIGGER brands_set_updated_at
    BEFORE UPDATE ON brands
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE categories (
    id         BIGSERIAL PRIMARY KEY,
    parent_id  BIGINT      REFERENCES categories (id) ON DELETE RESTRICT,
    slug       TEXT        NOT NULL,
    name_fa    TEXT        NOT NULL,
    position   INTEGER     NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX categories_slug_key ON categories (slug) WHERE deleted_at IS NULL;
CREATE INDEX categories_parent_idx ON categories (parent_id, position) WHERE deleted_at IS NULL;

CREATE TRIGGER categories_set_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE products (
    id           BIGSERIAL PRIMARY KEY,
    slug         TEXT        NOT NULL,
    name_fa      TEXT        NOT NULL,
    name_en      TEXT,
    -- Active ingredient, the strongest signal when matching items across sources.
    generic_name TEXT,
    dosage_form  TEXT,
    strength     TEXT,
    brand_id     BIGINT      REFERENCES brands (id) ON DELETE SET NULL,
    category_id  BIGINT      REFERENCES categories (id) ON DELETE SET NULL,
    image_url    TEXT,
    description  TEXT,
    -- Global trade number and the Iranian drug code, when the source provides them.
    gtin         TEXT,
    irc          TEXT,
    status       TEXT        NOT NULL DEFAULT 'needs_review'
        CHECK (status IN ('published', 'hidden', 'needs_review')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE UNIQUE INDEX products_slug_key ON products (slug) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX products_gtin_key ON products (gtin) WHERE gtin IS NOT NULL AND deleted_at IS NULL;
CREATE UNIQUE INDEX products_irc_key ON products (irc) WHERE irc IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX products_brand_idx ON products (brand_id) WHERE deleted_at IS NULL;
CREATE INDEX products_category_idx ON products (category_id) WHERE deleted_at IS NULL;
CREATE INDEX products_status_idx ON products (status) WHERE deleted_at IS NULL;

CREATE TRIGGER products_set_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
