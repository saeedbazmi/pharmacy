import Link from "next/link";

import { opsRequest } from "@/lib/ops/server";
import type { OpsProduct } from "@/lib/ops/types";

export default async function OpsProductsPage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string; status?: string; page?: string }>;
}) {
  const q = await searchParams;
  const params = new URLSearchParams();
  if (q.q) params.set("q", q.q);
  if (q.status) params.set("status", q.status);
  if (q.page) params.set("page", q.page);
  const body = await opsRequest<{ products: OpsProduct[]; total: number }>(
    `/api/v1/ops/products?${params.toString()}`,
  );
  return (
    <main className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-8">
      <h1 className="text-2xl font-bold text-primary-dark">کالاها</h1>
      <form className="flex gap-3" method="get">
        <input name="q" defaultValue={q.q} placeholder="جست‌وجوی نام" className="rounded-lg border border-border px-3 py-2" />
        <select name="status" defaultValue={q.status ?? ""} className="rounded-lg border border-border px-3 py-2">
          <option value="">همه وضعیت‌ها</option>
          <option value="published">منتشر</option>
          <option value="hidden">مخفی</option>
          <option value="needs_review">بازبینی</option>
        </select>
        <button className="rounded-lg border border-border px-3 py-2" type="submit">
          فیلتر
        </button>
      </form>
      <p className="text-sm text-muted">{body.total} کالا</p>
      <ul className="flex flex-col gap-2">
        {body.products.map((p) => (
          <li key={p.id} className="flex justify-between rounded-lg border border-border bg-surface px-4 py-2">
            <Link href={`/ops/products/${p.id}`} className="text-primary">
              {p.name_fa}
            </Link>
            <span className="text-sm text-muted">{p.status}</span>
          </li>
        ))}
      </ul>
    </main>
  );
}
