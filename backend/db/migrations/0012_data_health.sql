-- M5: withhold suspicious price jumps, score source credibility, and
-- keep health / click reports on indexed time ranges.

ALTER TABLE offers
    ADD COLUMN IF NOT EXISTS proposed_price_rial BIGINT,
    ADD COLUMN IF NOT EXISTS proposed_at         TIMESTAMPTZ;

ALTER TABLE offers
    ADD CONSTRAINT offers_proposed_price_nonneg
        CHECK (proposed_price_rial IS NULL OR proposed_price_rial >= 0);

ALTER TABLE data_sources
    ADD COLUMN IF NOT EXISTS price_reject_count INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS offers_suspicious_idx
    ON offers (proposed_at DESC NULLS LAST)
    WHERE status = 'suspicious';

-- Windowed success-rate queries filter by started_at across sources.
CREATE INDEX IF NOT EXISTS sync_runs_started_idx
    ON sync_runs (started_at DESC);

-- Click reports already have (pharmacy_id, clicked_at) and
-- (product_id, clicked_at). A covering range scan on clicked_at helps
-- the "all pharmacies in a window" aggregate without a seq scan.
CREATE INDEX IF NOT EXISTS redirect_clicks_clicked_at_idx
    ON redirect_clicks (clicked_at);
