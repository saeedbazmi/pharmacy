---
name: frontend-dev
description: >-
  Frontend engineering rules for the pharmacy price-comparison platform:
  Next.js 15 App Router, React Server Components, TypeScript, Tailwind CSS v4
  with theme tokens, RTL Persian UI, price formatting, SEO/JSON-LD, performance
  budgets, and accessibility. Use when writing, reviewing, or refactoring
  anything under web/, when building a page, component, form, admin panel
  screen, or when the user mentions Next.js, React, Tailwind, UI, RTL, SEO, or
  styling.
---

# توسعه‌دهنده فرانت‌اند

نقش تو: مهندس فرانت‌اند این پروژه. استک و پالت رنگ در `AGENT.md` قطعی است؛ این فایل می‌گوید **چطور کد بنویس**.

## پیش از هر تغییر

1. `AGENT.md` را بخوان (پالت رنگ، بودجه عملکرد، قواعد SEO و RTL).
2. مشخص کن صفحه‌ی عمومی است یا پنل (`app/(admin)`)؛ قواعد SEO و بودجه‌ی JS فقط برای صفحات عمومی سخت‌گیرانه است.
3. اگر فیچر در `features.md` نیست، نساز؛ سؤال بپرس.

---

## اصول برنامه‌نویسی (اجباری)

- **مسئولیت واحد:** هر کامپوننت یک کار. کامپوننتی که هم داده می‌گیرد، هم فیلتر را مدیریت می‌کند، هم جدول را رندر می‌کند، باید شکسته شود.
- **Server-first:** پیش‌فرض همه‌چیز Server Component است. `"use client"` استثناست و باید توجیه داشته باشد (تعامل، state محلی، یا API مرورگر).
- **کامپوزیشن بر پیکربندی:** به‌جای کامپوننتی با ۱۵ prop بولی، از `children` و کامپوننت‌های کوچک ترکیب‌شونده استفاده کن.
- **بالا بردن state تا حد لازم، نه بیشتر:** state در نزدیک‌ترین جای ممکن. state سراسری در فاز اول لازم نیست؛ فیلتر و مرتب‌سازی در **URL search params** زندگی می‌کنند (قابل اشتراک، قابل ایندکس، بدون state manager).
- **انتزاع بعد از سه بار تکرار:** کامپوننت مشترک را وقتی بساز که الگو ثابت شده باشد، نه از پیش.
- **مرز داده:** فراخوانی API فقط از `lib/api`. `fetch` پراکنده در کامپوننت‌ها ممنوع.
- **تایپ صریح:** `any` ممنوع؛ نوع پاسخ API در یک جا (`lib/api/types.ts`) تعریف و همه‌جا از آن استفاده می‌شود. TypeScript در حالت `strict`.
- **بدون منطق کسب‌وکار در UI:** «کم‌ترین قیمت» و «معتبر بودن قیمت» را بک‌اند تعیین می‌کند؛ فرانت فقط نمایش می‌دهد.

---

## Server یا Client Component

| نیاز | انتخاب |
|---|---|
| نمایش داده، SEO، محتوای اولیه | Server Component |
| فیلتر، مرتب‌سازی، جست‌وجو | Server Component + خواندن `searchParams` |
| کلیک، ورودی کاربر، مودال، انتخاب مقایسه | Client Component کوچک و برگ‌مانند |
| تغییر داده (علاقه‌مندی، هشدار قیمت، فرم پنل) | Server Action |

قاعده: `"use client"` را در **برگ‌های درخت** بگذار، نه در ریشه‌ی صفحه. یک دکمه‌ی تعاملی نباید کل صفحه را کلاینتی کند.

```tsx
// app/product/[slug]/page.tsx — Server Component
export default async function ProductPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  const product = await api.getProduct(slug); // از lib/api
  if (!product) notFound();

  return (
    <article>
      <ProductHeader product={product} />
      <OfferTable offers={product.offers} />   {/* سرور: SEO لازم دارد */}
      <CompareToggle productId={product.id} /> {/* کلاینت: فقط این برگ */}
    </article>
  );
}
```

---

## ساختار و نام‌گذاری

```
web/
├─ app/
│  ├─ (public)/            # صفحه اصلی، جست‌وجو، کالا، دسته‌بندی
│  ├─ (admin)/             # پنل داده و مدیر
│  ├─ go/[offerId]/        # مسیر ری‌دایرکت خرید
│  └─ layout.tsx           # dir="rtl" lang="fa" + فونت
├─ components/
│  ├─ ui/                  # پایه‌ها: Button, Badge, Table, Input
│  └─ product/, offer/     # کامپوننت‌های دامنه‌ای
├─ lib/
│  ├─ api/                 # کلاینت + تایپ پاسخ‌ها
│  ├─ format.ts            # قیمت، تاریخ، ارقام فارسی
│  └─ persian.ts           # نرمال‌سازی متن ورودی
└─ styles/theme.css        # توکن‌های رنگ (تنها منبع رنگ)
```

فایل کامپوننت `PascalCase.tsx`، هوک `useSomething.ts`، ابزار `camelCase.ts`. هر فایل یک export اصلی. کامپوننت بیش از ۱۵۰ خط نشانه‌ی نیاز به شکستن است.

---

## استایل و رنگ

- **هیچ کد رنگ hex در کامپوننت.** همه‌ی رنگ‌ها از توکن‌های `styles/theme.css` که از جدول پالت `AGENT.md` می‌آید: `--color-primary`, `--color-accent`, `--color-best-price`, `--color-danger`, ....
- نارنجی `--color-accent` **فقط** برای دکمه «خرید از داروخانه». هیچ جای دیگر.
- رنگ به‌تنهایی حامل معنا نباشد: «موجود/ناموجود» همیشه متن یا آیکن هم داشته باشد.
- برای RTL از خصیصه‌های منطقی استفاده کن: `ps-4`/`pe-4`, `ms-2`/`me-2`, `text-start`/`text-end`. `left`/`right` و `pl`/`pr` ممنوع (به‌جز جایی که واقعاً فیزیکی است، مثل جهت آیکن فلش).
- کلاس Tailwind را در JSX نگه دار؛ فایل CSS جدا فقط برای توکن‌ها و ریست. `styled-components` یا CSS-in-JS ممنوع.
- هیچ UI-Kit سنگینی نصب نمی‌شود. کامپوننت پایه را در `components/ui` کپی/بنویس.
- کنتراست حداقل AA (۴.۵:۱) برای متن؛ حالت focus همیشه دیده شود (`focus-visible`).

---

## فارسی و RTL

- `<html lang="fa" dir="rtl">` در `layout.tsx`. فونت Vazirmatn به‌صورت self-host با `next/font/local` و `display: swap`.
- **نمایش** قیمت با ارقام فارسی و جداکننده هزارگان؛ **ذخیره و ارسال** همیشه لاتین و بر حسب ریال (عدد صحیح).

```ts
// lib/format.ts
export const formatPrice = (rial: number) =>
  new Intl.NumberFormat("fa-IR").format(rial) + " ریال";
```

- ورودی کاربر پیش از ارسال نرمال‌سازی می‌شود (`lib/persian.ts`): ارقام فارسی/عربی → لاتین، `ي/ك` عربی → فارسی، یکسان‌سازی نیم‌فاصله، حذف نویسه‌های صفرعرض.
- تاریخ شمسی فقط در لایه نمایش با `Intl.DateTimeFormat("fa-IR")`؛ داده‌ی خام همیشه UTC می‌ماند.
- زمان تازگی قیمت را همیشه نشان بده («به‌روزرسانی ۱۲ دقیقه پیش») — این ستون اعتماد محصول است.

---

## SEO (صفحات عمومی)

- هر صفحه `generateMetadata` دارد: `title`, `description`, `alternates.canonical`, و OG فارسی.
- صفحه کالا `JSON-LD` نوع `Product` + `AggregateOffer` با کم‌ترین/بیش‌ترین قیمت و ارز `IRR` دارد.
- صفحه کالا و دسته‌بندی با ISR رندر می‌شوند (`revalidate` متناسب با نرخ تغییر قیمت)؛ رندر کاملاً کلاینتی برای این صفحات ممنوع.
- `sitemap.ts` و `robots.ts` از داده‌ی دیتابیس تولید می‌شوند.
- دکمه خرید یک `<a>` به `/go/{offerId}` با `rel="nofollow sponsored"` است؛ نه `onClick` جاوااسکریپتی، تا بدون JS هم کار کند.
- یک `<h1>` در هر صفحه، سلسله‌مراتب هدینگ درست، و `alt` معنادار برای تصویر کالا.

---

## عملکرد

بودجه (از `AGENT.md`): LCP زیر ۲.۵ ثانیه روی 4G، JS اولیه هر صفحه زیر ۱۵۰KB گزیپ‌شده.

- `next/image` با `width`/`height` مشخص برای همه تصاویر؛ تصویر کالای بالای صفحه `priority`.
- کامپوننت سنگین و غیرضروری در بار اول (نمودار تاریخچه قیمت، مودال مقایسه) با `dynamic()` و بدون SSR.
- هیچ کتابخانه‌ی تاریخ یا آیکن سنگینی وارد نشود؛ `Intl` و SVG درون‌خطی کافی است.
- fetch در Server Component با `revalidate`/`tags` کش شود؛ پس از تغییر داده در پنل، `revalidateTag` صدا زده شود.
- لیست بلند نتایج با صفحه‌بندی سروری. اسکرول بی‌نهایت در فاز اول نه.
- از layout shift پرهیز: برای جدول قیمت و کارت‌ها skeleton با ابعاد ثابت.

---

## حالت‌های UI

هر صفحه‌ای که داده می‌گیرد باید چهار حالت داشته باشد و هیچ‌کدام فراموش نشود:

1. **بارگذاری** — `loading.tsx` یا skeleton، نه اسپینر تمام‌صفحه.
2. **خطا** — `error.tsx` با پیام فارسی قابل فهم + دکمه تلاش مجدد؛ متن خطای فنی هرگز به کاربر نشان داده نشود.
3. **خالی** — «این کالا در هیچ داروخانه‌ای موجود نیست» با پیشنهاد جای‌گزین یا لینک جست‌وجو.
4. **موفق** — حالت اصلی.

---

## فرم و تغییر داده

- تغییر داده با Server Action؛ اعتبارسنجی هم سمت کلاینت (تجربه) و هم سمت سرور (امنیت) انجام می‌شود و **سرور مرجع است**.
- ورودی OTP: نوع `tel`, `inputMode="numeric"`, `autoComplete="one-time-code"`، و پیام خطای شمارنده‌ی محدودیت ارسال.
- وضعیت در حال ارسال با `useFormStatus`؛ دکمه در زمان ارسال غیرفعال شود تا ارسال تکراری رخ ندهد.
- هرگز توکن یا سشن در `localStorage` نگه‌داری نشود؛ احراز هویت با کوکی `HttpOnly` است.

---

## دسترس‌پذیری

- عنصر معنادار به‌جای `div` کلیک‌پذیر: دکمه `<button>`، لینک `<a>`.
- هر ورودی `<label>` متصل دارد. آیکن‌های تنها `aria-label` دارند.
- مودال: قفل focus، بستن با `Escape`، بازگشت focus به عنصر آغازگر.
- جدول قیمت با `<table>` واقعی و `<th scope>`؛ نه شبیه‌سازی با div.
- پیمایش کامل با کیبورد باید ممکن باشد.

---

## چک‌لیست پیش از پایان کار

```
- [ ] پیش‌فرض Server Component؛ "use client" فقط در برگ‌ها و با دلیل
- [ ] هیچ hex رنگی در کامپوننت؛ فقط توکن‌های theme.css
- [ ] نارنجی accent فقط روی دکمه خرید
- [ ] کلاس‌های RTL منطقی (ps/pe/ms/me/text-start)
- [ ] قیمت با formatPrice، ورودی کاربر نرمال‌سازی شده
- [ ] متادیتا و JSON-LD برای صفحه عمومی جدید
- [ ] چهار حالت بارگذاری/خطا/خالی/موفق پوشش داده شده
- [ ] دسترس‌پذیری: label، focus دیده‌شدنی، پیمایش کیبوردی
- [ ] بودجه JS و LCP نقض نشده؛ تصاویر با next/image
- [ ] بدون any، بدون وابستگی جدید غیرموجه، لینتر و tsc تمیز
```

## ممنوعه‌ها

`any`، `fetch` مستقیم در کامپوننت، رنگ hex درون JSX، `pl/pr/left/right` در چیدمان RTL، state manager سراسری، UI-Kit سنگین، CSS-in-JS، ذخیره توکن در `localStorage`، رندر کلاینتی صفحات SEO-محور، منطق کسب‌وکار (تعیین ارزان‌ترین قیمت) در فرانت، نام‌گذاری فارسی برای متغیر و فایل.
