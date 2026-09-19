-- Job queue, sync run log, match review queue, and a normalized name column
-- used by first-pass product matching.

CREATE TABLE jobs (
    id           BIGSERIAL PRIMARY KEY,
    kind         TEXT        NOT NULL,
    payload      JSONB       NOT NULL DEFAULT '{}'::jsonb,
    status       TEXT        NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'running', 'retry', 'succeeded', 'failed')),
    attempts     INTEGER     NOT NULL DEFAULT 0,
    max_attempts INTEGER     NOT NULL DEFAULT 5,
    run_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_error   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX jobs_claim_idx ON jobs (run_at) WHERE status IN ('pending', 'retry');
CREATE INDEX jobs_kind_status_idx ON jobs (kind, status);

CREATE TRIGGER jobs_set_updated_at
    BEFORE UPDATE ON jobs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE sync_runs (
    id          BIGSERIAL PRIMARY KEY,
    source_id   BIGINT      NOT NULL REFERENCES data_sources (id) ON DELETE CASCADE,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    status      TEXT        NOT NULL DEFAULT 'running'
        CHECK (status IN ('running', 'succeeded', 'failed')),
    ok_count    INTEGER     NOT NULL DEFAULT 0,
    fail_count  INTEGER     NOT NULL DEFAULT 0,
    error       TEXT
);

CREATE INDEX sync_runs_source_idx ON sync_runs (source_id, started_at DESC);

CREATE TABLE match_candidates (
    id                   BIGSERIAL PRIMARY KEY,
    source_item_id       BIGINT      NOT NULL REFERENCES source_items (id) ON DELETE CASCADE,
    suggested_product_id BIGINT      REFERENCES products (id) ON DELETE SET NULL,
    score                NUMERIC(5, 4),
    status               TEXT        NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'rejected')),
    reason               TEXT        NOT NULL DEFAULT '',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX match_candidates_pending_idx ON match_candidates (status, created_at)
    WHERE status = 'pending';
CREATE UNIQUE INDEX match_candidates_open_item_key ON match_candidates (source_item_id)
    WHERE status = 'pending';

CREATE TRIGGER match_candidates_set_updated_at
    BEFORE UPDATE ON match_candidates
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE products
    ADD COLUMN name_normalized TEXT NOT NULL DEFAULT '';

CREATE INDEX products_name_normalized_idx ON products (name_normalized)
    WHERE deleted_at IS NULL;
