import Link from "next/link";

import { DashboardChart } from "@/components/ops/DashboardChart";
import { opsRequest } from "@/lib/ops/server";
import type { AdminDashboard, OpsUser } from "@/lib/ops/types";
import { formatNumber } from "@/lib/format";

interface PageProps {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}

export default async function DashboardPage({ searchParams }: PageProps) {
  const me = await opsRequest<OpsUser>("/api/v1/ops/me");
  if (me.role !== "super_admin") {
    return (
      <main className="mx-auto max-w-6xl px-6 py-8">
        <h1 className="text-2xl font-bold">داشبورد</h1>
        <p className="mt-3 text-muted">فقط مدیر کل به این بخش دسترسی دارد.</p>
      </main>
    );
  }
  const raw = (await searchParams).days;
  const days = Number.parseInt(Array.isArray(raw) ? (raw[0] ?? "") : (raw ?? ""), 10);
  const range = Number.isFinite(days) && days > 0 ? days : 7;
  const dash = await opsRequest<AdminDashboard>(`/api/v1/admin/dashboard?days=${range}`);
  const freshnessHours = dash.coverage.median_freshness_seconds / 3600;

  return (
    <main className="mx-auto flex max-w-6xl flex-col gap-8 px-6 py-8">
      <h1 className="text-2xl font-bold text-primary-dark">داشبورد مدیر</h1>
      <form className="flex flex-wrap items-center gap-2 text-sm" method="get">
        <label>
          بازه
          <select
            name="days"
            defaultValue={String(range)}
            className="ms-2 rounded-lg border border-border px-3 py-2"
          >
            <option value="7">۷ روز</option>
            <option value="30">۳۰ روز</option>
            <option value="90">۹۰ روز</option>
          </select>
        </label>
        <button type="submit" className="rounded-lg border border-border px-3 py-2">
          نمایش
        </button>
      </form>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Stat label="جست‌وجو" value={formatNumber(dash.searches)} />
        <Stat label="کلیک ری‌دایرکت" value={formatNumber(dash.clicks)} />
        <Stat label="نرخ کلیک" value={formatCTR(dash.ctr)} />
        <Stat
          label="میانگین تازگی قیمت"
          value={`${formatNumber(Math.round(freshnessHours))} ساعت`}
        />
        <Stat label="داروخانه فعال" value={formatNumber(dash.coverage.active_pharmacies)} />
        <Stat label="کالای منتشرشده" value={formatNumber(dash.coverage.published_products)} />
      </div>
      <section className="rounded-xl border border-border bg-surface p-4">
        <h2 className="mb-3 text-lg font-medium">ترافیک روزانه</h2>
        <p className="mb-2 text-xs text-muted">میله روشن: جست‌وجو · میله تیره: کلیک</p>
        <DashboardChart series={dash.series ?? []} />
      </section>
      <section>
        <h2 className="mb-3 text-lg font-medium">کلیک به‌تفکیک داروخانه</h2>
        {dash.pharmacies.length === 0 ? (
          <p className="text-muted">در این بازه کلیکی ثبت نشده.</p>
        ) : (
          <ul className="flex flex-col gap-2">
            {dash.pharmacies.map((p) => (
              <li key={p.id} className="flex justify-between rounded-lg border border-border bg-surface px-4 py-2">
                <span>{p.name}</span>
                <span className="numeric">{formatNumber(p.clicks)}</span>
              </li>
            ))}
          </ul>
        )}
      </section>
      <div className="grid gap-6 lg:grid-cols-2">
        <TermList title="جست‌وجوهای پرتکرار" items={dash.frequent} />
        <TermList title="جست‌وجوهای بی‌نتیجه" items={dash.zero} />
      </div>
      <p className="text-sm text-muted">
        گزارش کلیک جزئی‌تر در{" "}
        <Link href="/ops/clicks" className="text-primary hover:text-primary-dark">
          گزارش کلیک
        </Link>{" "}
        است.
      </p>
    </main>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-border bg-surface p-4">
      <div className="text-sm text-muted">{label}</div>
      <div className="mt-1 text-2xl font-bold">{value}</div>
    </div>
  );
}

function formatCTR(ctr: number): string {
  return `${formatNumber(Math.round(ctr * 1000) / 10)}٪`;
}

function TermList({
  title,
  items,
}: {
  title: string;
  items: { query: string; hits: number }[];
}) {
  return (
    <section>
      <h2 className="mb-3 text-lg font-medium">{title}</h2>
      {items.length === 0 ? (
        <p className="text-muted">موردی نیست.</p>
      ) : (
        <ol className="flex flex-col gap-2">
          {items.slice(0, 20).map((item) => (
            <li
              key={item.query}
              className="flex justify-between rounded-lg border border-border bg-surface px-4 py-2 text-sm"
            >
              <span>{item.query}</span>
              <span className="numeric text-muted">{formatNumber(item.hits)}</span>
            </li>
          ))}
        </ol>
      )}
    </section>
  );
}
