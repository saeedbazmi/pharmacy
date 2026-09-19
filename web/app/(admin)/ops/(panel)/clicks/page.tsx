import { formatNumber } from "@/lib/format";
import { opsRequest } from "@/lib/ops/server";
import type { OpsClickReport } from "@/lib/ops/types";

export default async function ClicksPage({
  searchParams,
}: {
  searchParams: Promise<{ from?: string; to?: string }>;
}) {
  const q = await searchParams;
  const params = new URLSearchParams();
  if (q.from) params.set("from", q.from);
  if (q.to) params.set("to", q.to);
  const qs = params.toString();
  const report = await opsRequest<OpsClickReport>(`/api/v1/ops/reports/clicks${qs ? `?${qs}` : ""}`);
  const csvHref = `/api/v1/ops/reports/clicks.csv${qs ? `?${qs}` : ""}`;

  return (
    <main className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-8">
      <h1 className="text-2xl font-bold text-primary-dark">گزارش کلیک ری‌دایرکت</h1>
      <p className="text-muted">
        ورودی تصمیم انتخاب داروخانه پایلوت فاز دوم. بازه روی پارتیشن ماهانه
        <code className="mx-1">redirect_clicks</code> فیلتر می‌شود.
      </p>
      <form className="flex flex-wrap items-end gap-3 text-sm" method="get">
        <label className="flex flex-col gap-1">
          از
          <input name="from" type="date" defaultValue={q.from} className="rounded-lg border border-border px-3 py-2" />
        </label>
        <label className="flex flex-col gap-1">
          تا
          <input name="to" type="date" defaultValue={q.to} className="rounded-lg border border-border px-3 py-2" />
        </label>
        <button className="rounded-lg border border-border px-3 py-2" type="submit">
          نمایش
        </button>
        <a href={csvHref} className="rounded-lg border border-border px-3 py-2 text-primary hover:text-primary-dark">
          دریافت CSV
        </a>
      </form>
      <p className="text-sm text-muted">مجموع کلیک: {formatNumber(report.total)}</p>
      {report.top_pharmacy ? (
        <p>
          پرکلیک‌ترین داروخانه: <strong>{report.top_pharmacy.name}</strong> با{" "}
          {formatNumber(report.top_pharmacy.clicks)} کلیک
        </p>
      ) : null}

      <section>
        <h2 className="mb-2 text-xl font-bold">به تفکیک داروخانه</h2>
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-right text-muted">
              <th className="py-2">داروخانه</th>
              <th>کلیک</th>
            </tr>
          </thead>
          <tbody>
            {report.pharmacies.map((row) => (
              <tr key={row.id} className="border-b border-border">
                <td className="py-2">{row.name}</td>
                <td>{formatNumber(row.clicks)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>

      <section>
        <h2 className="mb-2 text-xl font-bold">پرکلیک‌ترین کالاها</h2>
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-right text-muted">
              <th className="py-2">کالا</th>
              <th>کلیک</th>
            </tr>
          </thead>
          <tbody>
            {report.top_products.map((row) => (
              <tr key={row.id} className="border-b border-border">
                <td className="py-2">{row.name_fa}</td>
                <td>{formatNumber(row.clicks)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </main>
  );
}
