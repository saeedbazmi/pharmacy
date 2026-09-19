import Link from "next/link";

import { formatRelativeTime } from "@/lib/format";
import { opsRequest } from "@/lib/ops/server";
import type { OpsHealthSource, OpsSyncRun } from "@/lib/ops/types";

export default async function SourceHealthPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const [health, runs] = await Promise.all([
    opsRequest<{ sources: OpsHealthSource[] }>("/api/v1/ops/health?days=7"),
    opsRequest<{ runs: OpsSyncRun[] }>(`/api/v1/ops/health/sources/${id}/runs`),
  ]);
  const src = health.sources.find((s) => String(s.id) === id);
  return (
    <main className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-8">
      <p>
        <Link href="/ops/health" className="text-sm text-primary hover:text-primary-dark">
          بازگشت به سلامت
        </Link>
      </p>
      <h1 className="text-2xl font-bold text-primary-dark">
        تاریخچه اجرا · {src?.pharmacy_name ?? `منبع ${id}`}
      </h1>
      {src?.last_error ? (
        <pre className="overflow-auto rounded-xl border border-danger bg-surface p-4 text-sm text-danger">
          {src.last_error}
        </pre>
      ) : null}
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-border text-right text-muted">
            <th className="py-2">شروع</th>
            <th>وضعیت</th>
            <th>موفق</th>
            <th>ناموفق</th>
            <th>خطا</th>
          </tr>
        </thead>
        <tbody>
          {runs.runs.map((run) => (
            <tr key={run.id} className="border-b border-border align-top">
              <td className="py-2">{formatRelativeTime(run.started_at)}</td>
              <td>{run.status === "succeeded" ? "موفق" : run.status === "failed" ? "ناموفق" : run.status}</td>
              <td>{run.ok_count}</td>
              <td>{run.fail_count}</td>
              <td className="max-w-md break-words text-danger">{run.error || "—"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </main>
  );
}
