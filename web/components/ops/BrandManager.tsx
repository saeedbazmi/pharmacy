"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/Button";
import type { OpsBrand } from "@/lib/ops/types";

export function BrandManager({ brands }: { brands: OpsBrand[] }) {
  const router = useRouter();
  const [error, setError] = useState("");

  async function create(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const response = await fetch("/api/v1/ops/brands", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({ name_fa: data.get("name_fa"), slug: data.get("slug") }),
    });
    if (!response.ok) {
      setError("ساخت برند ناموفق بود.");
      return;
    }
    event.currentTarget.reset();
    router.refresh();
  }

  async function merge(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!confirm("ادغام برند تکرارپذیر نیست مگر از سابقه. ادامه؟")) return;
    const data = new FormData(event.currentTarget);
    const from = Number(data.get("from_id"));
    const into = Number(data.get("into_id"));
    const response = await fetch(`/api/v1/ops/brands/${from}/merge`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({ into_id: into }),
    });
    if (!response.ok) {
      const err = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
      setError(err?.error?.message ?? "ادغام ناموفق بود.");
      return;
    }
    router.refresh();
  }

  return (
    <div className="flex flex-col gap-6">
      <form onSubmit={create} className="flex flex-col gap-3 rounded-xl border border-border bg-surface p-4">
        <input name="name_fa" required placeholder="نام برند" className="rounded-lg border border-border px-3 py-2" />
        <input name="slug" placeholder="slug" className="rounded-lg border border-border px-3 py-2" />
        <Button type="submit" size="sm">
          افزودن
        </Button>
      </form>
      <form onSubmit={merge} className="flex flex-col gap-3 rounded-xl border border-border bg-surface p-4">
        <p className="text-sm text-muted">ادغام برند تکراری: کالاها به برند مقصد منتقل می‌شوند.</p>
        <select name="from_id" required className="rounded-lg border border-border px-3 py-2">
          {brands.map((b) => (
            <option key={b.id} value={b.id}>
              از {b.name_fa}
            </option>
          ))}
        </select>
        <select name="into_id" required className="rounded-lg border border-border px-3 py-2">
          {brands.map((b) => (
            <option key={`into-${b.id}`} value={b.id}>
              به {b.name_fa}
            </option>
          ))}
        </select>
        <Button type="submit" size="sm" variant="secondary">
          ادغام
        </Button>
      </form>
      {error ? <p className="text-sm text-danger">{error}</p> : null}
      <ul className="flex flex-col gap-2">
        {brands.map((b) => (
          <li key={b.id} className="rounded-lg border border-border bg-surface px-3 py-2">
            {b.name_fa} <span className="text-xs text-muted">{b.slug}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
