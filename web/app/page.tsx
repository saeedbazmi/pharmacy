import { Badge } from "@/components/ui/Badge";

/**
 * Placeholder home page for M0. The real search-first home page arrives with
 * M3, and the product page with M1 (PH1-011).
 */
export default function HomePage() {
  return (
    <main className="mx-auto flex max-w-3xl flex-col gap-6 px-6 py-16">
      <div className="flex items-center gap-3">
        <h1 className="text-3xl font-bold text-primary-dark">مقایسه قیمت دارو</h1>
        <Badge tone="info">در حال ساخت</Badge>
      </div>

      <p className="text-muted">
        قیمت دارو و محصولات سلامت را در چند داروخانه مقایسه کنید و برای خرید به
        سایت همان داروخانه بروید. خرید و پرداخت در سایت داروخانه انجام می‌شود.
      </p>

      <section className="rounded-xl border border-border bg-surface p-6">
        <h2 className="mb-3 text-lg font-medium">وضعیت توسعه</h2>
        <p className="text-sm text-muted">
          این نسخه فقط زیرساخت فاز اول (M0) است: سرویس‌ها، دیتابیس، لاگ
          ساختاریافته و پایه‌های ظاهری. جست‌وجو و صفحه کالا در Milestoneهای بعدی
          اضافه می‌شوند.
        </p>
      </section>
    </main>
  );
}
