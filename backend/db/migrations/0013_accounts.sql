-- Optional public accounts: OTP login, favorites, price alerts, search history.

CREATE TABLE users (
    id         BIGSERIAL PRIMARY KEY,
    phone      TEXT        NOT NULL,
    role       TEXT        NOT NULL DEFAULT 'user' CHECK (role = 'user'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX users_phone_key ON users (phone) WHERE deleted_at IS NULL;

CREATE TRIGGER users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE otp_codes (
    id            BIGSERIAL PRIMARY KEY,
    phone         TEXT        NOT NULL,
    code_hash     TEXT        NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL,
    consumed_at   TIMESTAMPTZ,
    attempt_count INTEGER     NOT NULL DEFAULT 0,
    ip            TEXT        NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX otp_codes_phone_idx ON otp_codes (phone, created_at DESC);

CREATE TABLE otp_sends (
    id         BIGSERIAL PRIMARY KEY,
    phone      TEXT        NOT NULL,
    ip         TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX otp_sends_phone_idx ON otp_sends (phone, created_at DESC);
CREATE INDEX otp_sends_ip_idx ON otp_sends (ip, created_at DESC);

CREATE TABLE user_sessions (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX user_sessions_user_idx ON user_sessions (user_id);
CREATE INDEX user_sessions_expires_idx ON user_sessions (expires_at);

CREATE TABLE favorites (
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    product_id BIGINT      NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, product_id)
);

CREATE INDEX favorites_product_idx ON favorites (product_id);

CREATE TABLE price_alerts (
    id                 BIGSERIAL PRIMARY KEY,
    user_id            BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    product_id         BIGINT      NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    kind               TEXT        NOT NULL CHECK (kind IN ('price_drop', 'back_in_stock')),
    target_price_rial  BIGINT CHECK (target_price_rial IS NULL OR target_price_rial >= 0),
    baseline_price_rial BIGINT     NOT NULL DEFAULT 0,
    stock_was_available BOOLEAN    NOT NULL DEFAULT TRUE,
    status             TEXT        NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'paused')),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at         TIMESTAMPTZ
);

CREATE INDEX price_alerts_user_idx ON price_alerts (user_id) WHERE deleted_at IS NULL AND status = 'active';
CREATE INDEX price_alerts_eval_idx ON price_alerts (product_id) WHERE deleted_at IS NULL AND status = 'active';

CREATE TRIGGER price_alerts_set_updated_at
    BEFORE UPDATE ON price_alerts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE alert_deliveries (
    id         BIGSERIAL PRIMARY KEY,
    alert_id   BIGINT      NOT NULL REFERENCES price_alerts (id) ON DELETE CASCADE,
    event_key  TEXT        NOT NULL,
    status     TEXT        NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    attempts   INTEGER     NOT NULL DEFAULT 0,
    last_error TEXT,
    sent_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX alert_deliveries_event_key ON alert_deliveries (alert_id, event_key);

ALTER TABLE search_queries
    ADD COLUMN IF NOT EXISTS user_id BIGINT REFERENCES users (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS search_queries_user_idx
    ON search_queries (user_id, queried_at DESC)
    WHERE user_id IS NOT NULL;
