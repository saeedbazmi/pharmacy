import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";

import { clearSearches, deleteAlert, removeFavorite } from "@/app/(public)/account/actions";
import { SubmitButton } from "@/components/account/SubmitButton";
import { getMe, listAlerts, listFavorites, listSearches } from "@/lib/account/server";
import { formatPrice } from "@/lib/format";

export const metadata: Metadata = {
  title: "حساب کاربری",
  robots: { index: false, follow: false },
};

export default async function AccountPage() {
  const me = await getMe();
  if (!me) {
    redirect("/login?next=/account");
  }

  const [favorites, alerts, searches] = await Promise.all([
    listFavorites(),
    listAlerts(),
    listSearches(),
  ]);

  return (
    <main className="mx-auto flex max-w-4xl flex-col gap-10 px-6 py-10">
      <header className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-primary-dark">حساب کاربری</h1>
          <p className="mt-1 text-sm text-muted">شماره {me.phone_masked}</p>
        </div>
        <form action="/account/logout" method="post">
          <SubmitButton pendingLabel="خروج…">خروج</SubmitButton>
        </form>
      </header>

      <section className="flex flex-col gap-3">
        <h2 className="text-lg font-medium">علاقه‌مندی‌ها</h2>
        {favorites.length === 0 ? (
          <p className="text-sm text-muted">
            هنوز کالایی ذخیره نشده. از{" "}
            <Link href="/search" className="text-primary hover:text-primary-dark">
              جست‌وجو
            </Link>{" "}
            کالای مورد نظر را پیدا کنید.
          </p>
        ) : (
          <ul className="flex flex-col gap-2">
            {favorites.map((item) => (
              <li
                key={item.product_id}
                className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-border bg-surface px-4 py-3"
              >
                <Link href={`/product/${item.slug}`} className="font-medium text-primary-dark">
                  {item.name_fa}
                </Link>
                <div className="flex items-center gap-3 text-sm">
                  {item.lowest_price_rial > 0 ? (
                    <span>{formatPrice(item.lowest_price_rial)}</span>
                  ) : (
                    <span className="text-muted">بدون قیمت معتبر</span>
                  )}
                  <form action={removeFavorite.bind(null, item.product_id)}>
                    <SubmitButton pendingLabel="…" className="text-sm">
                      حذف
                    </SubmitButton>
                  </form>
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="flex flex-col gap-3">
        <h2 className="text-lg font-medium">هشدارها</h2>
        {alerts.length === 0 ? (
          <p className="text-sm text-muted">
            هشداری ثبت نشده. در صفحه کالا می‌توانید کاهش قیمت یا موجود شدن را دنبال کنید.
          </p>
        ) : (
          <ul className="flex flex-col gap-2">
            {alerts.map((item) => (
              <li
                key={item.id}
                className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-border bg-surface px-4 py-3"
              >
                <div className="flex flex-col">
                  <Link href={`/product/${item.product_slug}`} className="font-medium text-primary-dark">
                    {item.product_name}
                  </Link>
                  <span className="text-sm text-muted">
                    {item.kind === "back_in_stock" ? "موجود شدن" : "کاهش قیمت"}
                    {item.target_price_rial
                      ? ` — هدف ${formatPrice(item.target_price_rial)}`
                      : ""}
                  </span>
                </div>
                <form action={deleteAlert.bind(null, item.id)}>
                  <SubmitButton pendingLabel="…">حذف</SubmitButton>
                </form>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="flex flex-col gap-3">
        <div className="flex items-center justify-between gap-3">
          <h2 className="text-lg font-medium">تاریخچه جست‌وجو</h2>
          {searches.length > 0 ? (
            <form action={clearSearches}>
              <SubmitButton pendingLabel="در حال پاک‌کردن…">پاک‌کردن تاریخچه</SubmitButton>
            </form>
          ) : null}
        </div>
        {searches.length === 0 ? (
          <p className="text-sm text-muted">جست‌وجویی با این حساب ثبت نشده است.</p>
        ) : (
          <ul className="flex flex-col gap-1 text-sm">
            {searches.map((item) => (
              <li key={item.id}>
                <Link
                  href={`/search?q=${encodeURIComponent(item.query)}`}
                  className="text-primary hover:text-primary-dark"
                >
                  {item.query}
                </Link>
                <span className="text-muted"> — {item.results} نتیجه</span>
              </li>
            ))}
          </ul>
        )}
      </section>
    </main>
  );
}
