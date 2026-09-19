import { opsRequest } from "@/lib/ops/server";
import type { OpsAudit } from "@/lib/ops/types";

export default async function AuditPage({
  searchParams,
}: {
  searchParams: Promise<{ actor_id?: string; since?: string; until?: string }>;
}) {
  const q = await searchParams;
  const params = new URLSearchParams();
  if (q.actor_id) params.set("actor_id", q.actor_id);
  if (q.since) params.set("since", q.since);
  if (q.until) params.set("until", q.until);
  const body = await opsRequest<{ entries: OpsAudit[]; total: number }>(`/api/v1/ops/audit?${params.toString()}`);
  return (
    <main className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-8">
      <h1 className="text-2xl font-bold text-primary-dark">سابقه تغییرات</h1>
      <form className="flex flex-wrap gap-3 text-sm" method="get">
        <input name="actor_id" defaultValue={q.actor_id} placeholder="شناسه کاربر" className="rounded-lg border border-border px-3 py-2" />
        <input name="since" type="date" defaultValue={q.since} className="rounded-lg border border-border px-3 py-2" />
        <input name="until" type="date" defaultValue={q.until} className="rounded-lg border border-border px-3 py-2" />
        <button className="rounded-lg border border-border px-3 py-2" type="submit">
          فیلتر
        </button>
      </form>
      <p className="text-sm text-muted">{body.total} ردیف</p>
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-border text-right text-muted">
            <th className="py-2">زمان</th>
            <th>کاربر</th>
            <th>موجودیت</th>
            <th>عمل</th>
          </tr>
        </thead>
        <tbody>
          {body.entries.map((row) => (
            <tr key={row.id} className="border-b border-border">
              <td className="py-2">{new Date(row.created_at).toLocaleString("fa-IR")}</td>
              <td>{row.actor_name}</td>
              <td>
                {row.entity} #{row.entity_id}
              </td>
              <td>{row.action}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </main>
  );
}
