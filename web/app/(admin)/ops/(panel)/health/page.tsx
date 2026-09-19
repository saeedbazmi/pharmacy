import Link from "next/link";

import { Badge } from "@/components/ui/Badge";
import { formatNumber, formatRelativeTime } from "@/lib/format";
import { opsRequest } from "@/lib/ops/server";
import type { OpsHealthSource, OpsStaleSource } from "@/lib/ops/types";

export default async function HealthPage({
  searchParams,
}: {
  searchParams: Promise<{ days?: string }>;
}) {
  const q = await searchParams;
  const days = q.days && Number(q.days) > 0 ? q.days : "7";
  const [health, stale] = await Promise.all([
    opsRequest<{ sources: OpsHealthSource[]; window_hours: number }>(`/api/v1/ops/health?days=${days}`),
    opsRequest<{ sources: OpsStaleSource[] }>("/api/v1/ops/reports/stale"),
  ]);

  return (
    <main className="mx-auto flex max-w-6xl flex-col gap-8 px-6 py-8">
      <div>
        <h1 className="text-2xl font-bold text-primary-dark">سلامت همگام‌سازی</h1>
        <p className="mt-2 text-muted">
          وضعیت واقعی هر منبع، آخرین خطا، و منابعی که از زمان‌بندی خود عقب افتاده‌اند.
          نرخ موفقیت برای {days} روز اخیر است.
        </p>
      </div>
      <form className="flex flex-wrap gap-3 text-sm" method="get">
        <label className="flex items-center gap-2">
          بازه
          <select name="days" defaultValue={days} className="rounded-lg border border-border px-3 py-2">
            <option value="1">۱ روز</option>
            <option value="7">۷ روز</option>
            <option value="30">۳۰ روز</option>
          </select>
        </label>
        <button className="rounded-lg border border-border px-3 py-2" type="submit">
          نمایش
        </button>
      </form>
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-border text-right text-muted">
            <th className="py-2">داروخانه</th>
            <th>وضعیت</th>
            <th>آخرین اجرا</th>
            <th>مدت</th>
            <th>موفق / ناموفق</th>
            <th>نرخ موفقیت</th>
            <th>خطا</th>
          </tr>
        </thead>
        <tbody>
          {health.sources.map((src) => {
            const status = healthStatus(src);
            return (
              <tr key={src.id} className="border-b border-border align-top">
                <td className="py-3">
                  <Link href={`/ops/health/${src.id}`} className="font-medium text-primary hover:text-primary-dark">
                    {src.pharmacy_name}
                  </Link>
                  {!src.enabled ? <div className="text-xs text-muted">غیرفعال</div> : null}
                </td>
                <td>
                  <Badge tone={status.tone}>
                    <StatusIcon kind={status.kind} />
                    {status.label}
                  </Badge>
                  {src.low_trust ? (
                    <div className="mt-1">
                      <Badge tone="warning">کم‌اعتبار · {formatNumber(src.reject_count)} رد قیمت</Badge>
                    </div>
                  ) : null}
                </td>
                <td>{src.last_run_at ? formatRelativeTime(src.last_run_at) : "هنوز اجرا نشده"}</td>
                <td>{src.run_duration_ms ? `${formatNumber(src.run_duration_ms)} ms` : "—"}</td>
                <td>
                  {src.latest_run
                    ? `${formatNumber(src.latest_run.ok_count)} / ${formatNumber(src.latest_run.fail_count)}`
                    : "—"}
                </td>
                <td>{src.runs ? `${formatNumber(Math.round(src.success_rate * 100))}٪ از ${formatNumber(src.runs)} اجرا` : "—"}</td>
                <td className="max-w-xs break-words text-danger">{src.last_error || "—"}</td>
              </tr>
            );
          })}
        </tbody>
      </table>

      <section className="flex flex-col gap-3">
        <h2 className="text-xl font-bold">کهنه‌ترین منابع</h2>
        <p className="text-sm text-muted">قدیمی‌ترین offer فعال هر داروخانه. آستانه هشدار ۲۴ ساعت و حذف از مقایسه ۷۲ ساعت است.</p>
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-right text-muted">
              <th className="py-2">داروخانه</th>
              <th>قدیمی‌ترین قیمت</th>
              <th>سن (ساعت)</th>
              <th>تعداد offer</th>
            </tr>
          </thead>
          <tbody>
            {stale.sources.map((src) => (
              <tr key={src.id} className="border-b border-border">
                <td className="py-2">{src.pharmacy_name}</td>
                <td>{formatRelativeTime(src.oldest_seen)}</td>
                <td>{formatNumber(Math.round(src.age_hours))}</td>
                <td>{formatNumber(src.offer_count)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </main>
  );
}

function healthStatus(src: OpsHealthSource): { label: string; tone: "success" | "warning" | "danger" | "neutral"; kind: "ok" | "late" | "fail" | "idle" } {
  if (src.last_status === "failed") {
    return { label: "ناموفق", tone: "danger", kind: "fail" };
  }
  if (src.overdue) {
    return { label: "عقب‌افتاده از زمان‌بندی", tone: "warning", kind: "late" };
  }
  if (src.last_status === "succeeded") {
    return { label: "سالم", tone: "success", kind: "ok" };
  }
  return { label: "بدون اجرا", tone: "neutral", kind: "idle" };
}

function StatusIcon({ kind }: { kind: "ok" | "late" | "fail" | "idle" }) {
  const title = { ok: "سالم", late: "عقب‌افتاده", fail: "ناموفق", idle: "بدون اجرا" }[kind];
  return (
    <svg aria-hidden="true" viewBox="0 0 16 16" className="size-3.5 shrink-0" fill="none">
      <title>{title}</title>
      {kind === "ok" ? (
        <path d="M3 8.2 6.2 11.5 13 4.5" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" />
      ) : null}
      {kind === "late" ? (
        <>
          <circle cx="8" cy="8" r="6" stroke="currentColor" strokeWidth="1.5" />
          <path d="M8 4.5v4l2.5 1.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
        </>
      ) : null}
      {kind === "fail" ? (
        <path d="M4 4 12 12M12 4 4 12" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
      ) : null}
      {kind === "idle" ? <circle cx="8" cy="8" r="5" stroke="currentColor" strokeWidth="1.5" /> : null}
    </svg>
  );
}
