-- Pharmacies and the data source (API or crawler) that feeds each one.
-- Money is stored as BIGINT rial and timestamps as TIMESTAMPTZ in UTC.

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE pharmacies (
    id          BIGSERIAL PRIMARY KEY,
    slug        TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    site_domain TEXT        NOT NULL,
    logo_url    TEXT,
    status      TEXT        NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'paused', 'disabled')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX pharmacies_slug_key ON pharmacies (slug) WHERE deleted_at IS NULL;
CREATE INDEX pharmacies_status_idx ON pharmacies (status) WHERE deleted_at IS NULL;

CREATE TRIGGER pharmacies_set_updated_at
    BEFORE UPDATE ON pharmacies
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE data_sources (
    id                BIGSERIAL PRIMARY KEY,
    pharmacy_id       BIGINT      NOT NULL REFERENCES pharmacies (id) ON DELETE CASCADE,
    kind              TEXT        NOT NULL CHECK (kind IN ('api', 'crawler')),
    -- Fetcher specific settings: endpoints, selectors, pagination.
    config            JSONB       NOT NULL DEFAULT '{}'::jsonb,
    schedule_interval INTERVAL    NOT NULL DEFAULT INTERVAL '1 hour',
    is_enabled        BOOLEAN     NOT NULL DEFAULT TRUE,
    last_run_at       TIMESTAMPTZ,
    last_status       TEXT        CHECK (last_status IN ('succeeded', 'failed', 'running')),
    last_error        TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX data_sources_pharmacy_idx ON data_sources (pharmacy_id) WHERE deleted_at IS NULL;
-- Drives the scheduler query: which enabled source is due to run next.
CREATE INDEX data_sources_due_idx ON data_sources (is_enabled, last_run_at) WHERE deleted_at IS NULL;

CREATE TRIGGER data_sources_set_updated_at
    BEFORE UPDATE ON data_sources
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
