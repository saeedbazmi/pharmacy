# task — M0: اسکلت و زیرساخت

معیار پذیرش هر آیتم در `plan.md` کنار همین فایل است؛ این فایل فهرست کار اجرایی است. هر تسک باید در یک نشست قابل انجام باشد.

نشانه‌ها: `[ ]` انجام‌نشده، `[~]` در جریان، `[x]` تمام. نقش: **B** = `backend-dev`، **F** = `frontend-dev`، **P** = `project-manager`.

---

## PH1-001 — اسکلت ریپو و اجرای محلی · B

- [ ] 001.1 ساخت درخت پوشه‌ها دقیقاً مطابق بخش ۳ `AGENT.md` (`backend/`, `web/`, `deploy/`, `docs/`)
- [ ] 001.2 `backend/go.mod` با Go 1.23 و فقط وابستگی‌های مجاز `AGENT.md`
- [ ] 001.3 `docker-compose.yml` با سرویس‌های `db` (postgres:16)، `api`، `web`، `worker` و healthcheck برای `db`
- [ ] 001.4 Dockerfile چندمرحله‌ای برای `api`/`worker` (باینری استاتیک) و Dockerfile فرانت
- [ ] 001.5 `Makefile` با هدف‌های `up`، `down`، `logs`، `migrate`، `sqlc`، `worker`، `test`، `lint`
- [ ] 001.6 `deploy/.env.example` با همه متغیرها و توضیح هرکدام
- [ ] 001.7 `.gitignore` و `.dockerignore` (بدون ورود `.env` و باینری به ریپو)
- [ ] 001.8 ساخت `docs/status.md` و `docs/backlog.md` خالی اولیه

## PH1-002 — سرور HTTP و middlewareها · B

- [ ] 002.1 `platform/httpx`: نوشتن پاسخ JSON، خواندن و اعتبارسنجی body، سقف حجم body
- [ ] 002.2 middleware تولید `request_id` و قرار دادنش در `context` و هدر پاسخ
- [ ] 002.3 middleware recover: لاگ `ERROR` با stack و پاسخ ۵۰۰ با قالب استاندارد
- [ ] 002.4 middleware لاگ دسترسی با متد، مسیر، وضعیت و `duration_ms`
- [ ] 002.5 middleware timeout درخواست
- [ ] 002.6 `internal/http/router.go` با گروه‌های `/api/v1`، `/api/v1/ops`، `/api/v1/admin` (جای‌نگه‌دار)
- [ ] 002.7 `internal/http/errors.go`: مپینگ sentinel به کد HTTP با `errors.Is` و قالب واحد خطا
- [ ] 002.8 `/healthz` و `/readyz` (بررسی ping دیتابیس)
- [ ] 002.9 `cmd/api/main.go` با راه‌اندازی و graceful shutdown

## PH1-003 — لاگر ساختاریافته · B

- [ ] 003.1 `platform/logger`: `slog` با JSON handler و سطح از env
- [ ] 003.2 تزریق فیلد `service` (`api` / `worker`) در ریشه لاگر
- [ ] 003.3 helper استخراج `request_id` از `context` برای همه‌ی `*Context` لاگ‌ها
- [ ] 003.4 `maskPhone` و تست جدولی آن
- [ ] 003.5 بازبینی: هیچ مسیری توکن، کوکی، OTP یا شماره کامل لاگ نمی‌کند

## PH1-004 — config، دیتابیس و ابزار مایگریشن · B

- [ ] 004.1 `platform/config`: خواندن env، اعتبارسنجی و توقف با خطای شفاف در نبود متغیر لازم
- [ ] 004.2 `platform/postgres`: `pgxpool` با سقف اتصال، timeout و ping در راه‌اندازی
- [ ] 004.3 `cmd/migrate`: اجرای ترتیبی فایل‌ها در ترنزکشن + جدول `schema_migrations`
- [ ] 004.4 `make migrate` روی دیتابیس خالی و چاپ وضعیت نهایی

## PH1-005 — مایگریشن اسکیمای هسته · B

- [ ] 005.1 `0001_pharmacies_sources.sql`: `pharmacies`، `data_sources` (با `config JSONB`, `schedule`)
- [ ] 005.2 `0002_catalog.sql`: `brands`، `categories`، `products` (با `slug` یکتا و `status`)
- [ ] 005.3 `0003_source_items.sql`: `source_items` با `raw JSONB`، `external_id`، `fetched_at`
- [ ] 005.4 `0004_offers.sql`: `offers` با یکتایی `(product_id, pharmacy_id)` + `price_history`
- [ ] 005.5 افزودن ایندکس‌های لازم در همان فایل‌ها (`slug`، کلید offer، `last_seen_at`)
- [ ] 005.6 بازبینی قواعد: مبلغ `BIGINT` ریال، زمان `TIMESTAMPTZ`، `deleted_at` برای موجودیت محتوایی
- [ ] 005.7 تست یکپارچه: اجرای همه مایگریشن‌ها روی Postgres تست

## PH1-006 — راه‌اندازی sqlc · B

- [ ] 006.1 `sqlc.yaml` با engine postgresql و خروجی pgx/v5
- [ ] 006.2 `db/queries/product.sql` با `GetProductBySlug`
- [ ] 006.3 `make sqlc` و ورود کد تولیدشده به ریپو
- [ ] 006.4 استفاده از کوئری تولیدشده در `module/catalog/repository.go`
- [ ] 006.5 بازبینی: هیچ رشته SQL دستی در کد Go نیست

## PH1-007 — اسکلت فرانت با RTL و پالت · F

- [ ] 007.1 راه‌اندازی Next.js 15 (App Router) در `web/` با TypeScript در حالت `strict`
- [ ] 007.2 `app/layout.tsx` با `lang="fa"`، `dir="rtl"` و متادیتای پیش‌فرض
- [ ] 007.3 افزودن Vazirmatn (subset فارسی، `woff2`) در `public/fonts` و اتصال با `next/font/local`
- [ ] 007.4 نصب Tailwind v4 و ساخت `styles/theme.css` با همه توکن‌های رنگ جدول `AGENT.md`
- [ ] 007.5 کامپوننت‌های پایه در `components/ui`: `Button`، `Badge`، `Table`
- [ ] 007.6 صفحه داخلی نمایش توکن‌ها با `robots: noindex` برای بازبینی چشمی
- [ ] 007.7 تنظیم ESLint و Prettier؛ `tsc` و لینتر بدون خطا

## PH1-008 — کلاینت API فرانت · F

- [ ] 008.1 `lib/api/client.ts` با base URL از env، timeout و ارسال `request_id`
- [ ] 008.2 `lib/api/types.ts` با تایپ پاسخ‌ها و بدون `any`
- [ ] 008.3 نگاشت قالب خطای بک‌اند به پیام فارسی + نگه‌داشتن `request_id` برای پیگیری
- [ ] 008.4 `lib/format.ts`: `formatPrice` و `formatRelativeTime` با `Intl` فارسی
- [ ] 008.5 `lib/persian.ts`: نرمال‌سازی ورودی کاربر (ارقام، `ي/ك`، نیم‌فاصله، نویسه صفرعرض)

---

## ترتیب پیشنهادی

001 → 004 → 005 → 003 → 002 → 006، و به‌صورت موازی 007 → 008.

## بازبینی پایانی

- [ ] همه معیارهای پذیرش `plan.md` تیک خورده‌اند
- [ ] `make lint` و `make test` تمیزند
- [ ] `docs/status.md` با وضعیت M0 به‌روز شده است
