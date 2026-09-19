"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/Button";
import type { OpsPharmacy, OpsSource, PreviewItem } from "@/lib/ops/types";

export function SourceForm({
  pharmacies,
  source,
}: {
  pharmacies: OpsPharmacy[];
  source?: OpsSource;
}) {
  const router = useRouter();
  const cfg = source?.config ?? {};
  const [error, setError] = useState("");
  const [preview, setPreview] = useState<PreviewItem[]>([]);
  const [pending, setPending] = useState(false);

  function payload(form: HTMLFormElement) {
    const data = new FormData(form);
    const createPharmacy = data.get("new_pharmacy") === "on";
    return {
      createPharmacy,
      pharmacy: {
        name: String(data.get("pharmacy_name") ?? ""),
        slug: String(data.get("pharmacy_slug") ?? ""),
        site_domain: String(data.get("site_domain") ?? ""),
      },
      source: {
        pharmacy_id: Number(data.get("pharmacy_id") || 0),
        kind: String(data.get("kind") ?? "crawler"),
        schedule_interval: String(data.get("schedule_interval") ?? "1 hour"),
        enabled: true,
        config: {
          fetcher: String(data.get("fetcher") ?? ""),
          listing_url: String(data.get("listing_url") ?? ""),
          max_pages: 1,
          api_key: String(data.get("api_key") ?? ""),
        },
      },
    };
  }

  async function save(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError("");
    const body = payload(event.currentTarget);
    let pharmacyId = body.source.pharmacy_id;
    if (!source && body.createPharmacy) {
      const created = await fetch("/api/v1/ops/pharmacies", {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: JSON.stringify(body.pharmacy),
      });
      if (!created.ok) {
        setPending(false);
        setError("ساخت داروخانه ناموفق بود.");
        return;
      }
      const row = (await created.json()) as { id: number };
      pharmacyId = row.id;
    }
    const path = source ? `/api/v1/ops/sources/${source.id}` : "/api/v1/ops/sources";
    const response = await fetch(path, {
      method: source ? "PATCH" : "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({ ...body.source, pharmacy_id: pharmacyId }),
    });
    setPending(false);
    if (!response.ok) {
      const err = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
      setError(err?.error?.message ?? "ذخیره منبع ناموفق بود.");
      return;
    }
    const saved = (await response.json()) as { id: number };
    router.push(`/ops/sources/${saved.id}`);
    router.refresh();
  }

  async function test(event: React.MouseEvent<HTMLButtonElement>) {
    event.preventDefault();
    const form = event.currentTarget.form;
    if (!form) return;
    setPending(true);
    setError("");
    const body = payload(form);
    const path = source ? `/api/v1/ops/sources/${source.id}/test` : "/api/v1/ops/sources/test";
    const response = await fetch(path, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify(body.source),
    });
    setPending(false);
    if (!response.ok) {
      const err = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
      setError(err?.error?.message ?? "تست اتصال ناموفق بود.");
      return;
    }
    const result = (await response.json()) as { items: PreviewItem[] };
    setPreview(result.items ?? []);
  }

  return (
    <form onSubmit={save} className="flex flex-col gap-4 rounded-xl border border-border bg-surface p-6">
      {!source ? (
        <>
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" name="new_pharmacy" defaultChecked />
            داروخانه تازه بساز
          </label>
          <label className="text-sm">
            داروخانه موجود
            <select name="pharmacy_id" className="mt-1 w-full rounded-lg border border-border px-3 py-2">
              <option value="">—</option>
              {pharmacies.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          </label>
          <label className="text-sm">
            نام داروخانه تازه
            <input name="pharmacy_name" className="mt-1 w-full rounded-lg border border-border px-3 py-2" />
          </label>
          <label className="text-sm">
            شناسه لاتین
            <input name="pharmacy_slug" className="mt-1 w-full rounded-lg border border-border px-3 py-2" />
          </label>
          <label className="text-sm">
            دامنه سایت
            <input name="site_domain" className="mt-1 w-full rounded-lg border border-border px-3 py-2" />
          </label>
        </>
      ) : (
        <input type="hidden" name="pharmacy_id" value={source.pharmacy_id} />
      )}
      <label className="text-sm">
        نوع
        <select name="kind" defaultValue={source?.kind ?? "crawler"} className="mt-1 w-full rounded-lg border border-border px-3 py-2">
          <option value="crawler">خزنده</option>
          <option value="api">API</option>
        </select>
      </label>
      <label className="text-sm">
        پیاده‌سازی موجود
        <select
          name="fetcher"
          defaultValue={String(cfg.fetcher ?? "darukade")}
          className="mt-1 w-full rounded-lg border border-border px-3 py-2"
        >
          <option value="darukade">داروکده</option>
          <option value="rosha">روشا</option>
        </select>
      </label>
      <label className="text-sm">
        نشانی فهرست
        <input
          name="listing_url"
          required
          defaultValue={String(cfg.listing_url ?? "")}
          className="mt-1 w-full rounded-lg border border-border px-3 py-2"
        />
      </label>
      <label className="text-sm">
        کلید API (فقط نوشتنی)
        <input name="api_key" placeholder={source ? "********" : ""} className="mt-1 w-full rounded-lg border border-border px-3 py-2" />
      </label>
      <label className="text-sm">
        بازه سینک
        <select
          name="schedule_interval"
          defaultValue={source?.schedule_interval ?? "1 hour"}
          className="mt-1 w-full rounded-lg border border-border px-3 py-2"
        >
          <option value="15 minutes">۱۵ دقیقه</option>
          <option value="30 minutes">۳۰ دقیقه</option>
          <option value="1 hour">۱ ساعت</option>
          <option value="6 hours">۶ ساعت</option>
          <option value="24 hours">۲۴ ساعت</option>
        </select>
      </label>
      {error ? <p className="text-sm text-danger">{error}</p> : null}
      <div className="flex gap-3">
        <Button type="submit" disabled={pending}>
          ذخیره
        </Button>
        <Button type="button" variant="secondary" onClick={test} disabled={pending}>
          تست اتصال
        </Button>
      </div>
      {preview.length > 0 ? (
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-right text-muted">
              <th className="py-2">نام</th>
              <th>برند</th>
              <th>قیمت تومان</th>
              <th>موجودی</th>
            </tr>
          </thead>
          <tbody>
            {preview.map((item) => (
              <tr key={item.external_id} className="border-b border-border">
                <td className="py-2">{item.name_fa}</td>
                <td>{item.brand_name}</td>
                <td className="numeric">{item.price_toman}</td>
                <td>{item.in_stock ? "هست" : "نیست"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : null}
    </form>
  );
}
