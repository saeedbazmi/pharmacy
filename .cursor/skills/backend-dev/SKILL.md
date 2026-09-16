---
name: backend-dev
description: >-
  Backend engineering rules for the pharmacy price-comparison platform: Go
  modular monolith, layering (handler/service/repository), pgx + sqlc queries,
  migrations, structured slog logging, job queue, crawler/ingestion pipeline,
  and testing standards. Use when writing, reviewing, or refactoring anything
  under backend/, when adding an API endpoint, database migration, sqlc query,
  data source or crawler, background job, or when the user mentions Go, API,
  Postgres, sqlc, worker, ingest, or product matching.
---

# توسعه‌دهنده بک‌اند

نقش تو: مهندس بک‌اند این پروژه. تصمیم‌های استک و معماری در `AGENT.md` قطعی است؛ این فایل می‌گوید **چطور کد بنویس**.

## پیش از هر تغییر

1. `AGENT.md` را بخوان (استک، مدل داده، قواعد لاگ).
2. مشخص کن تغییر به کدام module تعلق دارد. اگر جای مشخصی ندارد، به‌جای ساختن پوشه‌ی جدید، اول بپرس.
3. اگر فیچر در `features.md` نیست، پیاده‌سازی نکن؛ سؤال بپرس.

---

## اصول برنامه‌نویسی (اجباری)

این‌ها شعار نیستند؛ در ریویو بر اساسشان کد رد می‌شود.

- **مسئولیت واحد (SRP):** هر فایل و هر تابع یک دلیل تغییر داشته باشد. تابعی که هم HTTP پارس می‌کند، هم قیمت را اعتبارسنجی می‌کند، هم SQL می‌زند، باید شکسته شود.
- **وابستگی به انتزاع (DIP):** سرویس به interface وابسته است، نه به `pgxpool` یا کلاینت HTTP مشخص. interface را **در سمت مصرف‌کننده** تعریف کن، نه کنار پیاده‌سازی.
- **باز برای توسعه (OCP):** افزودن داروخانه‌ی جدید نباید هیچ فایل هسته‌ای را تغییر دهد؛ فقط یک `Fetcher` جدید. اگر مجبور شدی `switch` روی نوع داروخانه بزنی، طراحی غلط است.
- **interface کوچک (ISP):** interface یک‌تا سه متد. `Repository` غول‌آسا با ۲۰ متد ممنوع؛ به‌تفکیک نیاز بشکن (`ProductReader`, `OfferWriter`).
- **YAGNI:** انتزاع برای «شاید فردا» نساز. مرزهای لازم در `AGENT.md` تعیین شده‌اند؛ بیشتر از آن نه.
- **صراحت بر زیرکی:** کد قابل خواندن مقدم بر کد کوتاه. reflection، `interface{}`/`any` و generic غیرضروری ممنوع.
- **DRY با احتیاط:** دو کد شبیه با دو دلیل تغییر متفاوت را یکی نکن. تکرار ارزان‌تر از انتزاع اشتباه است.
- **بدون حالت جهانی:** هیچ متغیر global، هیچ singleton، هیچ `init()` که کار انجام دهد. همه‌چیز با constructor تزریق می‌شود.
- **مرز ماژول:** یک module هرگز `repository` ماژول دیگر را import نمی‌کند؛ فقط interface سرویسش را.

---

## لایه‌ها: چه چیزی کجا

| لایه | فایل | مسئول | ممنوع |
|---|---|---|---|
| Transport | `handler.go` | decode، اعتبارسنجی شکل ورودی، فراخوانی سرویس، مپینگ خطا به HTTP | هر قاعده کسب‌وکار، هر SQL |
| Domain | `service.go` | قاعده کسب‌وکار، ترنزکشن، هم‌ارکستری | شناخت HTTP (بدون `http.Request`)، ساخت SQL |
| Data | `repository.go` | اجرای کوئری sqlc، مپ به تایپ دامنه | تصمیم کسب‌وکار، محاسبه قیمت |
| Model | `model.go` | تایپ دامنه | تگ `json`، وابستگی به pgx |

DTO ورودی/خروجی HTTP در لایه‌ی transport می‌ماند؛ تایپ دامنه هرگز مستقیم به JSON سریالایز نمی‌شود.

### الگوی استاندارد یک module

```go
// module/pricing/repository.go — interface در سمت مصرف‌کننده
type OfferRepository interface {
    ListByProduct(ctx context.Context, productID int64) ([]Offer, error)
    Upsert(ctx context.Context, o Offer) error
}

// module/pricing/service.go
type Service struct {
    offers OfferRepository
    log    *slog.Logger
}

func NewService(offers OfferRepository, log *slog.Logger) *Service {
    return &Service{offers: offers, log: log}
}

func (s *Service) OffersForProduct(ctx context.Context, productID int64) ([]Offer, error) {
    offers, err := s.offers.ListByProduct(ctx, productID)
    if err != nil {
        return nil, fmt.Errorf("list offers for product %d: %w", productID, err)
    }
    return rankByPrice(offers), nil // قاعده کسب‌وکار، تست‌پذیر و خالص
}
```

قاعده: منطق قابل تست را در تابع **خالص** (بدون I/O) جدا کن (`rankByPrice`, `normalizeName`, `isSuspiciousJump`). این توابع باید بدون دیتابیس تست شوند.

---

## خطا

- خطای دامنه به‌صورت sentinel در همان module: `var ErrProductNotFound = errors.New("product not found")`.
- همه‌جا wrap با context معنادار: `fmt.Errorf("upsert offer %d: %w", id, err)`. پیام خطا حرف کوچک و بدون علامت پایانی.
- **یا wrap کن یا log کن، نه هر دو.** لاگ خطا فقط در بالاترین لایه (handler یا اجراکننده job).
- `err` هرگز نادیده گرفته نشود. اگر واقعاً بی‌اهمیت است، با `_ = f()` و یک دلیل کوتاه.
- panic فقط برای خطای برنامه‌نویسی در زمان راه‌اندازی. هیچ panic در مسیر درخواست؛ middleware recover لازم است اما نباید به آن تکیه شود.
- مپینگ به HTTP فقط در یک جای متمرکز (`internal/http/errors.go`) با `errors.Is`:

```json
{"error":{"code":"product_not_found","message":"کالا یافت نشد","request_id":"..."}}
```

پیام خطای بازگشتی به کاربر فارسی و بدون جزئیات داخلی است؛ جزئیات فقط در لاگ.

---

## context و مهلت زمانی

- `ctx` اولین پارامتر هر تابع I/O است. هرگز `context.Background()` در مسیر درخواست.
- هر فراخوانی بیرونی timeout صریح دارد: DB حداکثر ۳ ثانیه، HTTP منابع خارجی حداکثر ۱۵ ثانیه.
- `http.Client` مشترک با `Timeout` مشخص بساز؛ `http.DefaultClient` ممنوع (بدون timeout است).
- goroutine بدون مالک و بدون راه توقف ممنوع. برای کار موازی از `errgroup` با `SetLimit` استفاده کن.

---

## دیتابیس

- هر کوئری در `db/queries/*.sql` با نام‌گذاری sqlc نوشته می‌شود؛ سپس `make sqlc`. رشته SQL دستی در Go ممنوع.
- مایگریشن: فایل جدید و شماره‌گذاری‌شده در `db/migrations`، فقط forward. ویرایش مایگریشن مرج‌شده ممنوع.
- ایندکس همراه با همان مایگریشنی که کوئریِ نیازمندش را می‌آورد اضافه شود: `slug`، `(product_id, pharmacy_id)`، ستون‌های trigram جست‌وجو، و کلیدهای زمانیِ گزارش‌ها.
- مبلغ `BIGINT` ریال، هرگز float. زمان `TIMESTAMPTZ` در UTC؛ تبدیل به تقویم شمسی فقط در فرانت.
- ترنزکشن در **سرویس** آغاز می‌شود، نه در repository:

```go
func (s *Service) ApproveMatch(ctx context.Context, id int64) error {
    return s.tx.InTx(ctx, func(ctx context.Context) error {
        if err := s.matches.Approve(ctx, id); err != nil { return err }
        return s.offers.RelinkFromMatch(ctx, id)
    })
}
```

- به‌روزرسانی دسته‌جمعی قیمت با batch/`COPY` انجام شود، نه حلقه‌ی تک‌ردیفی. هر عملیات روی مجموعه، N+1 نداشته باشد.
- حذف نرم (`deleted_at`) برای موجودیت محتوایی؛ هر کوئری خواندن باید آن را فیلتر کند.

---

## لاگ

```go
s.log.InfoContext(ctx, "offer.updated",
    "source_id", src.ID, "product_id", p.ID, "price", price, "duration_ms", ms)
```

- نام رویداد `domain.action` با حروف کوچک: `search.query`, `redirect.click`, `crawl.failed`, `match.auto_approved`, `sync.completed`.
- `request_id` در middleware ساخته و در `ctx` جریان دارد؛ همیشه `InfoContext`/`ErrorContext` استفاده کن تا خودکار ضمیمه شود.
- سطح: `INFO` رویداد کسب‌وکار، `WARN` خطای قابل بازیابی (retry کرال)، `ERROR` نیازمند دخالت انسان. `DEBUG` در محیط عملیاتی خاموش.
- **هرگز** شماره موبایل کامل، کد OTP، توکن، کوکی یا هدر `Authorization` را لاگ نکن. موبایل ماسک شود (`0912***4567`).
- لاگ در حلقه‌ی داغ (هر آیتم کرال) ممنوع؛ به‌جایش خلاصه‌ی هر اجرا (`sync.completed` با شمارش موفق/ناموفق).
- رویدادهای گزارش‌شدنی (کلیک ری‌دایرکت، جست‌وجو) علاوه بر لاگ در جدول Postgres ثبت می‌شوند؛ ثبت کلیک غیرهم‌زمان است و هرگز ری‌دایرکت را کند نمی‌کند.

---

## صف کار و ingestion

- Job با `SELECT ... FOR UPDATE SKIP LOCKED` برداشته می‌شود؛ هر job باید **idempotent** باشد (اجرای دوباره داده را خراب نکند).
- هر job: `attempts`, `max_attempts`, backoff نمایی، و در شکست نهایی وضعیت `failed` با پیام خطا. هرگز retry بی‌نهایت.
- خطای یک منبع نباید بقیه را متوقف کند؛ هر منبع در job مستقل.
- افزودن داروخانه جدید = پیاده‌سازی `Fetcher` + یک ردیف `data_sources`. هیچ فایل هسته‌ای تغییر نمی‌کند:

```go
type Fetcher interface {
    Fetch(ctx context.Context, src Source) ([]RawItem, error)
}
```

- کرال محترمانه: rate-limit به‌تفکیک دامنه، احترام به `robots.txt`، User-Agent شفاف، backoff، timeout.
- پاسخ خام هر منبع در `source_items.raw` ذخیره می‌شود؛ نرمال‌سازی و تطبیق روی داده‌ی ذخیره‌شده انجام می‌شود تا قابل بازپردازش باشد.
- نرمال‌سازی فارسی پیش از تطبیق: `ي/ك` عربی، نیم‌فاصله، ارقام فارسی/عربی → لاتین، حذف نویسه‌های صفرعرض. یک تابع مشترک در `platform`، نه کپی در هر fetcher.
- تطبیق: کد یکتا (GTIN/IRC) → نام نرمال‌شده‌ی دقیق → شباهت trigram با آستانه. زیر آستانه → `match_candidates` برای تایید دستی، **نه** ادغام خودکار.
- جهش قیمت بیش از ±۷۰٪ → offer `suspicious` و صف بازبینی، نه انتشار.

---

## امنیت

- ورودی در مرز سنجیده می‌شود: طول، بازه، فرمت، مقادیر مجاز enum. هیچ ورودی خارجی مورد اعتماد نیست.
- URL منابع کرال اعتبارسنجی شود (فقط `http`/`https`، مسدودسازی IP داخلی) تا SSRF ممکن نشود.
- ری‌دایرکت خرید فقط به `product_url` ذخیره‌شده در دیتابیس؛ هرگز به URL دلخواه از query string (open redirect).
- OTP: طول عمر کوتاه، محدودیت تعداد ارسال به‌ازای شماره و IP، ذخیره‌ی هش‌شده، مقایسه با زمان ثابت، باطل‌شدن پس از یک بار استفاده.
- کوکی سشن `HttpOnly` + `Secure` + `SameSite=Lax`. بررسی نقش (`data_ops`, `super_admin`) در middleware مسیرهای `/ops` و `/admin`.
- هر تغییر تیم داده و مدیر در `audit_logs` با actor و مقدار قبل/بعد ثبت شود.
- هیچ راز در کد یا ریپو؛ فقط env و مستندسازی در `deploy/.env.example`.

---

## تست

- `service`: تست واحد با repository جعلی (پیاده‌سازی ساده‌ی interface، نه کتابخانه mock).
- توابع خالص (تطبیق، نرمال‌سازی، رتبه‌بندی قیمت، تشخیص جهش): تست جدولی با موارد مرزی. **بدون تست مرج نمی‌شوند.**
- هر `Fetcher`: تست روی HTML/JSON واقعی ذخیره‌شده در `testdata/`. بدون تماس شبکه در تست.
- repository و مایگریشن: تست یکپارچه روی Postgres واقعی (کانتینر یا DB تست)، نه SQLite.
- تست را از رفتار بنویس نه از پیاده‌سازی؛ اسم تست بگوید چه چیزی تضمین می‌شود.

---

## چک‌لیست پیش از پایان کار

```
- [ ] لایه‌بندی رعایت شده (handler نازک، منطق در service، repository بی‌منطق)
- [ ] هیچ import مستقیم repository ماژول دیگر
- [ ] ctx و timeout در همه مسیرهای I/O
- [ ] خطاها wrap شده‌اند و مپینگ HTTP متمرکز است
- [ ] کوئری از sqlc آمده و ایندکس لازم در مایگریشن هست
- [ ] لاگ با نام رویداد domain.action و بدون داده حساس
- [ ] job جدید idempotent و با سقف retry است
- [ ] تست واحد برای منطق جدید نوشته شده
- [ ] go vet و gofmt و لینتر تمیز، بدون وابستگی جدید غیرموجه
```

## ممنوعه‌ها

ORM، رشته SQL دستی، global state، `panic` در مسیر درخواست، `http.DefaultClient`، goroutine بی‌مالک، لاگ داده حساس، ویرایش مایگریشن قدیمی، افزودن وابستگی جدید بدون توجیه، نام‌گذاری فارسی برای متغیر/تابع/جدول.
