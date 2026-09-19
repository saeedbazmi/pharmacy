-- Click log for the buy redirect. Partitioned by month so reports in later
-- milestones can drop or archive old ranges without touching live inserts.

CREATE TABLE redirect_clicks (
    id          BIGSERIAL,
    clicked_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    offer_id    BIGINT      NOT NULL,
    product_id  BIGINT      NOT NULL,
    pharmacy_id BIGINT      NOT NULL,
    referrer    TEXT,
    PRIMARY KEY (id, clicked_at)
) PARTITION BY RANGE (clicked_at);

CREATE INDEX redirect_clicks_pharmacy_idx ON redirect_clicks (pharmacy_id, clicked_at DESC);
CREATE INDEX redirect_clicks_offer_idx ON redirect_clicks (offer_id, clicked_at DESC);
CREATE INDEX redirect_clicks_product_idx ON redirect_clicks (product_id, clicked_at DESC);

CREATE OR REPLACE FUNCTION ensure_redirect_clicks_month(target date)
    RETURNS void
    LANGUAGE plpgsql
AS
$$
DECLARE
    start_at date := date_trunc('month', target::timestamp)::date;
    end_at   date := (date_trunc('month', target::timestamp) + INTERVAL '1 month')::date;
    part     text := 'redirect_clicks_' || to_char(start_at, 'YYYY_MM');
BEGIN
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF redirect_clicks FOR VALUES FROM (%L) TO (%L)',
        part, start_at, end_at
    );
END;
$$;

SELECT ensure_redirect_clicks_month(CURRENT_DATE);
SELECT ensure_redirect_clicks_month((CURRENT_DATE + INTERVAL '1 month')::date);
