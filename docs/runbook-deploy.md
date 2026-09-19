# استقرار و بازگشت — فاز اول

تاریخ: ۱۹ سپتامبر ۲۰۲۶

این رویه برای میزبان لینوکس با Docker Compose است. رازها فقط از `.env` می‌آیند؛ `deploy/.env.example` فهرست کامل کلیدهاست.

## پیش‌نیاز

- دامنهٔ عمومی روی A/AAAA به همین سرور اشاره کند.
- پورت ۸۰ و ۴۴۳ از بیرون باز باشد. پورت ۵۴۳۲، ۸۰۸۰ و ۸۰۸۱ را به اینترنت باز نگذارید.
- `.env` با `APP_ENV=production`، `OTP_PRINT_CODE=false`، `OTP_PEPPER` حداقل ۱۶ نویسه، `POSTGRES_PASSWORD` قوی و `NEXT_PUBLIC_SITE_URL=https://<دامنه>`.

## استقرار

1. کد را روی تگ/کامیت مشخص چک‌اوت کنید.
2. `cp deploy/.env.example .env` و مقادیر عملیاتی را پر کنید. `SITE_DOMAIN` و `CADDY_EMAIL` برای TLS لازم‌اند.
3. مایگریشن **قبل از** نسخهٔ جدید API اجرا می‌شود چون سرویس `migrate` در compose به `api` و `worker` با `service_completed_successfully` وابسته است:
   ```bash
   docker compose up -d --build db migrate
   docker compose up -d --build api worker web
   docker compose --profile launch up -d caddy
   ```
   یا یکجا: `docker compose --profile launch up -d --build`.
4. سلامت:
   - `https://<دامنه>/healthz` و `/readyz` → api
   - `https://<دامنه>/healthz/worker` و `/readyz/worker` → worker
5. اولین مدیر کل: `OPS_ROLE=super_admin make create-ops-user`

Caddy برای دامنهٔ واقعی گواهی Let’s Encrypt می‌گیرد. روی `localhost` فقط HTTP است.

## پشتیبان

- روزانه: `0 3 * * * /opt/pharmacy/deploy/backup.sh` با `DATABASE_URL` یا compose.
- خروجی: `deploy/backups/pharmacy-<utc>.sql.gz`؛ نگه‌داری ۱۴ روز.
- بازیابی: `CONFIRM=yes BACKUP=deploy/backups/<file>.sql.gz ./deploy/restore.sh`
- پشتیبان بدون آزمون بازیابی پشتیبان نیست. پس از اولین استقرار، یک‌بار روی نمونهٔ تازه بازیابی کنید و `/readyz` را ببینید.

## لاگ

سرویس‌ها JSON روی stdout می‌نویسند. درایور `json-file` با سقف ۲۰MB × ۷ فایل است.

جست‌وجو:

```bash
docker compose logs api --since 24h | jq -r 'select(.msg=="redirect.click")'
docker compose logs api --since 1h | jq -r 'select(.msg=="search.query")'
```

شماره موبایل کامل، OTP، توکن و کوکی در لاگ نیست.

## بازگشت (rollback)

1. کامیت قبلی را چک‌اوت کنید.
2. **مایگریشن‌ها فقط جلو می‌روند.** اگر نسخهٔ جدید جدول ساخته، برگشت کد بدون برگشت طرح‌واره ممکن است. در آن حالت از آخرین پشتیبان سالم بازیابی کنید.
3. دوباره `docker compose up -d --build` (مایگریشن no-op اگر از قبل اعمال شده).
4. `/readyz` و یک جست‌وجو و یک `/go/{id}` را دستی بزنید.

## امنیت مسیرها

- `/api/v1/ops/*` و `/api/v1/admin/*` بدون `ops_session` پاسخ ۴۰۱ می‌دهند.
- `/api/v1/admin/*` فقط `super_admin` است؛ `data_ops` و نقش `user` ۴۰۳ می‌گیرند.
- نقش `user` کوکی `user_session` دارد و پنل را باز نمی‌کند.
