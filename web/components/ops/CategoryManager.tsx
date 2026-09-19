"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/Button";
import type { OpsCategory } from "@/lib/ops/types";

export function CategoryManager({ categories }: { categories: OpsCategory[] }) {
  const router = useRouter();
  const [error, setError] = useState("");

  async function create(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const response = await fetch("/api/v1/ops/categories", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({
        name_fa: data.get("name_fa"),
        slug: data.get("slug"),
        parent_id: Number(data.get("parent_id") || 0),
      }),
    });
    if (!response.ok) {
      setError("ساخت دسته ناموفق بود.");
      return;
    }
    event.currentTarget.reset();
    router.refresh();
  }

  async function disable(id: number) {
    const reassign = prompt("اگر کالا دارد، شناسه دسته مقصد را بنویسید؛ وگرنه خالی بگذارید.");
    if (reassign === null) return;
    const response = await fetch(`/api/v1/ops/categories/${id}/disable`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({ reassign_to: Number(reassign || 0) }),
    });
    if (!response.ok) {
      const err = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
      setError(err?.error?.message ?? "حذف ممکن نیست.");
      return;
    }
    router.refresh();
  }

  const roots = categories.filter((c) => !c.parent_id);
  const childrenOf = (id: number) => categories.filter((c) => c.parent_id === id);

  return (
    <div className="flex flex-col gap-6">
      <form onSubmit={create} className="flex flex-col gap-3 rounded-xl border border-border bg-surface p-4">
        <input name="name_fa" required placeholder="نام دسته" className="rounded-lg border border-border px-3 py-2" />
        <input name="slug" placeholder="slug" className="rounded-lg border border-border px-3 py-2" />
        <select name="parent_id" className="rounded-lg border border-border px-3 py-2">
          <option value="">بدون والد</option>
          {categories.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name_fa}
            </option>
          ))}
        </select>
        <Button type="submit" size="sm">
          افزودن
        </Button>
      </form>
      {error ? <p className="text-sm text-danger">{error}</p> : null}
      <ul className="flex flex-col gap-2">
        {roots.map((c) => (
          <li key={c.id}>
            <Node category={c} onDisable={disable} />
            <ul className="mr-6 mt-1 flex flex-col gap-1">
              {childrenOf(c.id).map((child) => (
                <li key={child.id}>
                  <Node category={child} onDisable={disable} />
                </li>
              ))}
            </ul>
          </li>
        ))}
      </ul>
    </div>
  );
}

function Node({ category, onDisable }: { category: OpsCategory; onDisable: (id: number) => void }) {
  return (
    <div className="flex items-center justify-between rounded-lg border border-border bg-surface px-3 py-2">
      <span>
        {category.name_fa} <span className="text-xs text-muted">{category.slug}</span>
      </span>
      <Button size="sm" variant="ghost" type="button" onClick={() => onDisable(category.id)}>
        حذف
      </Button>
    </div>
  );
}
