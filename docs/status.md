# وضعیت — ۱۷ سپتامبر ۲۰۲۶

**Milestone فعلی:** M0 — اسکلت و زیرساخت
**نتیجه هدف:** سرویس‌ها بالا بیایند، دیتابیس مایگریت شود، لاگ JSON خروجی داشته باشد

## تمام‌شده

- **PH1-001** اسکلت ریپو، `docker-compose.yml`، `Makefile`، `deploy/.env.example`، `.dockerignore`
- **PH1-002** سرور HTTP با روتر استاندارد، middlewareهای `request_id`/recover/access-log/timeout/max-body، `/healthz`، `/readyz`، مپینگ متمرکز خطا
- **PH1-003** لاگر `slog` با خروجی JSON، فیلدهای پایه، `MaskPhone` و تست
- **PH1-004** `config` با اعتبارسنجی در راه‌اندازی، `pgxpool`، `cmd/migrate` با `schema_migrations`
- **PH1-005** مایگریشن اسکیمای هسته: `pharmacies`، `data_sources`، `brands`، `categories`، `products`، `source_items`، `offers`، `price_history`
- **PH1-006** `sqlc` با خروجی pgx/v5 و کوئری `GetProductBySlug`
- **PH1-007** اسکلت Next.js 15 با RTL، فونت Vazirmatn (self-host)، `styles/theme.css`، کامپوننت‌های پایه، صفحه بازبینی توکن‌ها
- **PH1-008** کلاینت `lib/api` با تایپ پاسخ‌ها، مپینگ خطا، `lib/format.ts`، `lib/persian.ts`

## در جریان

- تایید دروازه خروج M0 روی دیتابیس واقعی: اجرای `make migrate` و سبز شدن `/readyz`. Docker Desktop تازه نصب شده و ممکن است به راه‌اندازی مجدد سیستم نیاز داشته باشد.

## بلوکه

- ندارد.

## اندازه‌گیری‌ها

- تست بک‌اند: سبز (`internal/http`، `module/catalog`، `platform/config`، `platform/logger`).
- `go vet` و `next lint` بدون خطا.
- JS اولیه صفحه‌های عمومی: **۱۰۳KB** (بودجه: زیر ۱۵۰KB) — با حجم واقعی محتوا در M2 دوباره سنجیده می‌شود.

## تصمیم‌های لازم از کاربر

1. کدام داروخانه منبع اول M1 است و آیا API دارد یا باید کرال شود؟
2. بازه زمانی سینک منبع اول چقدر باشد؟
3. دامنه نهایی و محیط استقرار (برای M7، فعلاً جای‌نگه‌دار).

## انحراف‌های ثبت‌شده از برنامه

- کد تولیدی `sqlc` در `backend/internal/platform/store` قرار گرفت (در `AGENT.md` مسیر صریحی برایش تعیین نشده بود؛ `platform` جای زیرساخت است).
- برای اثبات کامل مسیر `router → service → repository → sqlc → db`، endpoint عمومی `GET /api/v1/products/{slug}` در M0 پیاده شد. این بخشی از PH1-013.4 در M1 است و همان‌جا با offer تکمیل می‌شود.
- کامنت‌های کد و فایل‌های تنظیمات انگلیسی نوشته شدند (طبق قاعده ۷ `AGENT.md`)؛ فارسی فقط در متن UI و مستندات.
