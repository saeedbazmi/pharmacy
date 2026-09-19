"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { revalidateCatalog } from "@/app/(admin)/ops/actions";
import { Button } from "@/components/ui/Button";
import type { OpsBrand, OpsCategory, OpsProduct } from "@/lib/ops/types";

export function ProductEditForm({
  product,
  brands,
  categories,
}: {
  product: OpsProduct;
  brands: OpsBrand[];
  categories: OpsCategory[];
}) {
  const router = useRouter();
  const snap = product.source_snapshot ?? {};
  const [error, setError] = useState("");

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const status = String(data.get("status"));
    if (status === "hidden" && !confirm("کالا از سایت و جست‌وجو مخفی شود؟")) return;
    const response = await fetch(`/api/v1/ops/products/${product.id}`, {
      method: "PATCH",
      credentials: "include",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({
        name_fa: data.get("name_fa"),
        name_en: data.get("name_en"),
        generic_name: data.get("generic_name"),
        image_url: data.get("image_url"),
        brand_id: Number(data.get("brand_id") || 0),
        category_id: Number(data.get("category_id") || 0),
        status,
      }),
    });
    if (!response.ok) {
      const err = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
      setError(err?.error?.message ?? "ذخیره ناموفق بود.");
      return;
    }
    setError("");
    await revalidateCatalog(product.slug);
    router.refresh();
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-4">
      <p className="text-sm text-muted">
        مقدار منبع کنار فیلد دستی است. فیلد ویرایش‌شده قفل می‌شود تا سینک بعدی آن را
        بازنویسی نکند.
      </p>
      <Field name="name_fa" label="نام فارسی" value={product.name_fa} source={String(snap.name_fa ?? "")} locked={product.locked_fields.includes("name_fa")} />
      <Field name="name_en" label="نام انگلیسی" value={product.name_en ?? ""} source={String(snap.name_en ?? "")} locked={product.locked_fields.includes("name_en")} />
      <Field name="generic_name" label="ماده مؤثره" value={product.generic_name ?? ""} source="" locked={product.locked_fields.includes("generic_name")} />
      <Field name="image_url" label="تصویر" value={product.image_url ?? ""} source={String(snap.image_url ?? "")} locked={product.locked_fields.includes("image_url")} />
      <label className="text-sm">
        برند
        <select name="brand_id" defaultValue={product.brand_id ?? ""} className="mt-1 w-full rounded-lg border border-border px-3 py-2">
          <option value="">—</option>
          {brands.map((b) => (
            <option key={b.id} value={b.id}>
              {b.name_fa}
            </option>
          ))}
        </select>
      </label>
      <label className="text-sm">
        دسته
        <select name="category_id" defaultValue={product.category_id ?? ""} className="mt-1 w-full rounded-lg border border-border px-3 py-2">
          <option value="">—</option>
          {categories.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name_fa}
            </option>
          ))}
        </select>
      </label>
      <label className="text-sm">
        وضعیت
        <select name="status" defaultValue={product.status} className="mt-1 w-full rounded-lg border border-border px-3 py-2">
          <option value="published">منتشر</option>
          <option value="hidden">مخفی</option>
          <option value="needs_review">بازبینی</option>
        </select>
      </label>
      {error ? <p className="text-sm text-danger">{error}</p> : null}
      <Button type="submit">ذخیره</Button>
    </form>
  );
}

function Field({
  name,
  label,
  value,
  source,
  locked,
}: {
  name: string;
  label: string;
  value: string;
  source: string;
  locked: boolean;
}) {
  return (
    <label className="text-sm">
      {label} {locked ? <span className="text-warning">قفل</span> : null}
      <input name={name} defaultValue={value} className="mt-1 w-full rounded-lg border border-border px-3 py-2" />
      {source ? <span className="mt-1 block text-xs text-muted">مقدار منبع: {source}</span> : null}
    </label>
  );
}
