import Link from "next/link";

import { opsRequest } from "@/lib/ops/server";
import type { OpsSource } from "@/lib/ops/types";

export default async function SourcesPage() {
  const body = await opsRequest<{ sources: OpsSource[] }>("/api/v1/ops/sources");
  return (
    <main className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-primary-dark">منابع داده</h1>
        <Link href="/ops/sources/new" className="rounded-lg bg-primary px-4 py-2 text-sm text-white">
          منبع تازه
        </Link>
      </div>
      <table className="w-full border-collapse text-sm">
        <thead>
          <tr className="border-b border-border text-right text-muted">
            <th className="py-2">داروخانه</th>
            <th>نوع</th>
            <th>زمان‌بند</th>
            <th>آخرین اجرا</th>
            <th>وضعیت</th>
          </tr>
        </thead>
        <tbody>
          {body.sources.map((src) => (
            <tr key={src.id} className="border-b border-border">
              <td className="py-2">
                <Link href={`/ops/sources/${src.id}`} className="text-primary hover:text-primary-dark">
                  {src.pharmacy_name}
                </Link>
              </td>
              <td>{src.kind}</td>
              <td>{src.schedule_interval}</td>
              <td>{src.last_status || "—"}</td>
              <td>{src.enabled ? "فعال" : "غیرفعال"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </main>
  );
}
