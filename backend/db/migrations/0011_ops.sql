-- Internal operators, sessions, audit trail, field locks, and category slug history.

CREATE TABLE internal_users (
    id            BIGSERIAL PRIMARY KEY,
    username      TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    role          TEXT        NOT NULL CHECK (role IN ('data_ops', 'super_admin')),
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

CREATE UNIQUE INDEX internal_users_username_key ON internal_users (username) WHERE deleted_at IS NULL;

CREATE TRIGGER internal_users_set_updated_at
    BEFORE UPDATE ON internal_users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE internal_sessions (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES internal_users (id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX internal_sessions_user_idx ON internal_sessions (user_id, expires_at);

CREATE TABLE login_attempts (
    id         BIGSERIAL PRIMARY KEY,
    username   TEXT        NOT NULL,
    ip         TEXT        NOT NULL,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX login_attempts_lookup_idx ON login_attempts (username, ip, attempted_at DESC);

CREATE TABLE audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    actor_id    BIGINT      REFERENCES internal_users (id) ON DELETE SET NULL,
    actor_name  TEXT        NOT NULL,
    entity      TEXT        NOT NULL,
    entity_id   TEXT        NOT NULL,
    action      TEXT        NOT NULL,
    before_data JSONB       NOT NULL DEFAULT '{}'::jsonb,
    after_data  JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX audit_logs_actor_time_idx ON audit_logs (actor_id, created_at DESC);
CREATE INDEX audit_logs_time_idx ON audit_logs (created_at DESC);
CREATE INDEX audit_logs_entity_idx ON audit_logs (entity, entity_id);

CREATE TABLE category_slug_redirects (
    old_slug    TEXT        PRIMARY KEY,
    category_id BIGINT      NOT NULL REFERENCES categories (id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS locked_fields TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS source_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE match_candidates
    ADD COLUMN IF NOT EXISTS decided_by BIGINT REFERENCES internal_users (id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS decided_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS previous_product_id BIGINT REFERENCES products (id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS linked_product_id BIGINT REFERENCES products (id) ON DELETE SET NULL;
