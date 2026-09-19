# شاخص‌ها — ۱۹ سپتامبر ۲۰۲۶

پوشش فعلی پس از M7. منابع: داروکده و روشا. پنل داخلی روی `/ops`. حساب اختیاری روی `/login`. داشبورد مدیر روی `/ops/dashboard`.

| شاخص | مقدار | یادداشت |
|---|---:|---|
| داروخانه فعال | ۲ | `darukade`، `rosha` |
| کالا (`products`) | ۳۰٬۰۷۳ | شامل bulk جست‌وجو |
| offer فعال | ۷۲ | مشکوک از سایت حذف است |
| آستانه جهش قیمت | ۰.۷۰ | `PRICE_JUMP_RATIO`؛ ADR-002 |
| کهنه / بحرانی | ۲۴س / ۷۲س | `OFFER_STALE_AFTER` / `OFFER_CRITICAL_AFTER` |
| صف قیمت مشکوک | `/ops/prices` | KPI بازبینی |
| رد قیمت منبع | `data_sources.price_reject_count` | ≥۳ = کم‌اعتبار |
| نرخ موفقیت sync | `/ops/health?days=7` | از `sync_runs.started_at` |
| کلیک به‌تفکیک داروخانه | `/ops/clicks` و `/ops/dashboard` | ورودی پایلوت فاز دوم |
| ترافیک جست‌وجو / CTR | `/ops/dashboard?days=7` | `search_queries` ÷ `redirect_clicks` |
| `match_candidates` تاییدشده | ۵۰ | از M4 |
| آستانه شباهت جست‌وجو | ۰.۲۸ | `search.MinSimilarity` |
| صدک ۹۵ جست‌وجو | ۲۵ms | از سنجش M3؛ بودجه ۲۰۰ms |
| JS اولیهٔ مشترک (gzip) | ۱۰۳KB | First Load JS shared؛ بودجه ۱۵۰KB |
| JS صفحه اصلی | ۱۱۲KB | بیلد Next M7 |
| JS صفحه کالا | ۱۱۴KB | `/product/[slug]` |
| JS نتایج جست‌وجو | ۱۱۳KB | `/search` |
| LCP هدف | <۲٫۵s روی 4G | ISR ۶۰s، `next/image` priority، فونت swap |
| LCP مشاهده‌شدهٔ محلی | زیر بودجه در شبکهٔ توسعه | شبیه‌سازی 4G را روی دامنهٔ واقعی با Lighthouse تکرار کنید |
| عمر OTP | ۲ دقیقه | هش‌شده؛ یک‌بارمصرف |
| سقف ارسال OTP | ۵ / شماره و ۱۰ / IP در ۱۵د | ADR-003 |
| عمر سشن کاربر | ۳۰ روز | کوکی `user_session` |
| سقف هشدار فعال | ۲۰ | هر کاربر |
| فلگ خرید مستقیم | خاموش | کش ۳۰s؛ بدون اثر کاربری |

صفحه نمونه:

- http://localhost:3000/
- http://localhost:3000/about
- http://localhost:3000/search?q=%D8%A7%D8%B3%D8%AA%D8%A7%D9%85%DB%8C%D9%86%D9%88%D9%81%D9%86
- http://localhost:3000/ops/dashboard
- http://localhost:3000/ops/users
- http://localhost:3000/ops/settings
- http://localhost:3000/ops/flags
- http://localhost:3000/login
- http://localhost:3000/ops/login
