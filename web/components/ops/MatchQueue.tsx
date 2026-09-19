"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";

import { Button } from "@/components/ui/Button";
import type { OpsMatch, OpsSource } from "@/lib/ops/types";

export function MatchQueue({
  matches,
  total,
  page,
  sources,
  sourceId,
  minScore,
}: {
  matches: OpsMatch[];
  total: number;
  page: number;
  sources: OpsSource[];
  sourceId: string;
  minScore: string;
}) {
  const router = useRouter();
  const [linkId, setLinkId] = useState<number | null>(null);
  const [productId, setProductId] = useState("");
  const [error, setError] = useState("");

  async function act(id: number, action: string, body?: unknown) {
    if (action === "reject" && !confirm("رد شود؟")) return;
    const response = await fetch(`/api/v1/ops/matches/${id}/${action}`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: body ? JSON.stringify(body) : "{}",
    });
    if (!response.ok) {
      const err = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
      setError(err?.error?.message ?? "عملیات ناموفق بود.");
      return;
    }
    setLinkId(null);
    router.refresh();
  }

  return (
    <div className="flex flex-col gap-4">
      <form className="flex flex-wrap gap-3 text-sm" method="get">
        <select name="source_id" defaultValue={sourceId} className="rounded-lg border border-border px-3 py-2">
          <option value="">همه منابع</option>
          {sources.map((src) => (
            <option key={src.id} value={src.id}>
              {src.pharmacy_name}
            </option>
          ))}
        </select>
        <input
          name="min_score"
          defaultValue={minScore}
          placeholder="حداقل امتیاز"
          className="rounded-lg border border-border px-3 py-2"
        />
        <Button type="submit" size="sm" variant="secondary">
          فیلتر
        </Button>
      </form>
      <p className="text-sm text-muted">{total} مورد در انتظار</p>
      {error ? <p className="text-sm text-danger">{error}</p> : null}
      {matches.map((match) => (
        <article key={match.id} className="grid gap-4 rounded-xl border border-border bg-surface p-4 md:grid-cols-2">
          <div>
            <h2 className="font-medium">داده منبع #{match.external_id}</h2>
            <pre className="mt-2 max-h-56 overflow-auto rounded-lg bg-background p-3 text-xs">
              {JSON.stringify(match.raw, null, 2)}
            </pre>
          </div>
          <div className="flex flex-col gap-3">
            <div>
              <h2 className="font-medium">کالای پیشنهادی</h2>
              {match.suggested_name ? (
                <p>
                  {match.suggested_name}{" "}
                  {match.suggested_slug ? (
                    <Link href={`/product/${match.suggested_slug}`} className="text-primary">
                      مشاهده عمومی
                    </Link>
                  ) : null}
                </p>
              ) : (
                <p className="text-muted">پیشنهادی نیست</p>
              )}
              <p className="text-sm text-muted">امتیاز {match.score}</p>
            </div>
            <div className="flex flex-wrap gap-2">
              <Button size="sm" onClick={() => act(match.id, "approve")} disabled={!match.suggested_product_id}>
                تایید
              </Button>
              <Button size="sm" variant="secondary" onClick={() => act(match.id, "reject")}>
                رد
              </Button>
              <Button size="sm" variant="secondary" onClick={() => act(match.id, "create-product")}>
                کالای تازه
              </Button>
              <Button size="sm" variant="ghost" onClick={() => setLinkId(match.id)}>
                اتصال دستی
              </Button>
            </div>
            {linkId === match.id ? (
              <div className="flex gap-2">
                <input
                  value={productId}
                  onChange={(e) => setProductId(e.target.value)}
                  placeholder="شناسه کالا"
                  className="rounded-lg border border-border px-3 py-1"
                />
                <Button size="sm" onClick={() => act(match.id, "link", { product_id: Number(productId) })}>
                  اتصال
                </Button>
              </div>
            ) : null}
          </div>
        </article>
      ))}
      <div className="flex gap-3 text-sm">
        {page > 1 ? (
          <Link href={`/ops/matches?page=${page - 1}`} className="text-primary">
            قبلی
          </Link>
        ) : null}
        <Link href={`/ops/matches?page=${page + 1}`} className="text-primary">
          بعدی
        </Link>
      </div>
    </div>
  );
}
