-- Launch readiness: public site settings, feature flags, dashboard indexes.

CREATE TABLE site_settings (
    key        TEXT        PRIMARY KEY,
    value      JSONB       NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER site_settings_set_updated_at
    BEFORE UPDATE ON site_settings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

INSERT INTO site_settings (key, value) VALUES
(
    'seo',
    '{"title":"مقایسه قیمت دارو","description":"قیمت دارو و محصولات سلامت را در چند داروخانه آنلاین مقایسه کنید و برای خرید به سایت همان داروخانه بروید."}'::jsonb
),
(
    'banner',
    '{"enabled":false,"title":"","body":""}'::jsonb
),
(
    'featured_categories',
    '{"slugs":[]}'::jsonb
),
(
    'pages',
    $${
      "about": "این وب‌سایت قیمت دارو و محصولات سلامت را از داروخانه‌های آنلاین جمع می‌کند تا بتوانید مقایسه کنید و برای خرید به سایت همان داروخانه بروید. ما فروشنده نیستیم و سفارشی ثبت نمی‌کنیم.",
      "contact": "برای گزارش خطای داده یا درخواست منبع جدید، از طریق ایمیل عملیاتی اعلام‌شده در استقرار پیام بفرستید. این نشانی برای مشاوره درمانی نیست.",
      "terms": "استفاده از این وب‌سایت به معنای پذیرش این قواعد است: اطلاعات قیمت از منابع داروخانه‌ها می‌آید و ممکن است تا چند ساعت کهنه باشد. خرید و پرداخت فقط در سایت داروخانه انجام می‌شود. پلتفرم واسطه فروش، سبد خرید یا پرداخت نیست.",
      "disclaimer": "این وب‌سایت مرجع تشخیص یا درمان نیست و توصیه پزشکی یا دارویی ارائه نمی‌دهد. تصمیم درمانی را با پزشک یا داروساز بگیرید. قیمت و موجودی نهایی فقط در سایت داروخانه معتبر است."
    }$$::jsonb
);

CREATE TABLE feature_flags (
    id          BIGSERIAL PRIMARY KEY,
    key         TEXT        NOT NULL,
    pharmacy_id BIGINT      REFERENCES pharmacies (id) ON DELETE CASCADE,
    enabled     BOOLEAN     NOT NULL DEFAULT FALSE,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX feature_flags_global_key
    ON feature_flags (key)
    WHERE pharmacy_id IS NULL;

CREATE UNIQUE INDEX feature_flags_pharmacy_key
    ON feature_flags (key, pharmacy_id)
    WHERE pharmacy_id IS NOT NULL;

CREATE TRIGGER feature_flags_set_updated_at
    BEFORE UPDATE ON feature_flags
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

INSERT INTO feature_flags (key, pharmacy_id, enabled)
VALUES ('direct_purchase', NULL, FALSE);

-- Dashboard freshness: active offers scanned by last_seen_at.
CREATE INDEX IF NOT EXISTS offers_active_last_seen_idx
    ON offers (last_seen_at)
    WHERE status = 'active';
