"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { revalidateSite } from "@/app/(admin)/ops/actions";
import { Button } from "@/components/ui/Button";
import type { SiteSettings } from "@/lib/site";

export function SettingsForm({ settings }: { settings: SiteSettings }) {
  const router = useRouter();
  const [error, setError] = useState("");

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const body: SiteSettings = {
      seo: {
        title: String(data.get("seo_title") ?? ""),
        description: String(data.get("seo_description") ?? ""),
      },
      banner: {
        enabled: data.get("banner_enabled") === "on",
        title: String(data.get("banner_title") ?? ""),
        body: String(data.get("banner_body") ?? ""),
      },
      featured_categories: {
        slugs: String(data.get("featured_slugs") ?? "")
          .split(/[,\s]+/)
          .map((s) => s.trim())
          .filter(Boolean),
      },
      pages: {
        about: String(data.get("about") ?? ""),
        contact: String(data.get("contact") ?? ""),
        terms: String(data.get("terms") ?? ""),
        disclaimer: String(data.get("disclaimer") ?? ""),
      },
    };
    const response = await fetch("/api/v1/admin/settings", {
      method: "PUT",
      credentials: "include",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify(body),
    });
    if (!response.ok) {
      const err = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
      setError(err?.error?.message ?? "ذخیره تنظیمات ناموفق بود.");
      return;
    }
    setError("");
    await revalidateSite();
    router.refresh();
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-6">
      {error ? <p className="text-sm text-danger">{error}</p> : null}
      <fieldset className="flex flex-col gap-3 rounded-xl border border-border bg-surface p-4">
        <legend className="px-1 font-medium">SEO</legend>
        <label className="text-sm">
          عنوان پیش‌فرض
          <input name="seo_title" defaultValue={settings.seo.title} className="mt-1 w-full rounded-lg border border-border px-3 py-2" />
        </label>
        <label className="text-sm">
          توضیح پیش‌فرض
          <textarea name="seo_description" defaultValue={settings.seo.description} rows={3} className="mt-1 w-full rounded-lg border border-border px-3 py-2" />
        </label>
      </fieldset>
      <fieldset className="flex flex-col gap-3 rounded-xl border border-border bg-surface p-4">
        <legend className="px-1 font-medium">بنر صفحه اصلی</legend>
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" name="banner_enabled" defaultChecked={settings.banner.enabled} />
          نمایش بنر
        </label>
        <label className="text-sm">
          عنوان بنر
          <input name="banner_title" defaultValue={settings.banner.title} className="mt-1 w-full rounded-lg border border-border px-3 py-2" />
        </label>
        <label className="text-sm">
          متن بنر
          <textarea name="banner_body" defaultValue={settings.banner.body} rows={3} className="mt-1 w-full rounded-lg border border-border px-3 py-2" />
        </label>
      </fieldset>
      <label className="text-sm">
        اسلاگ دسته‌های ویژه (با فاصله یا ویرگول)
        <input
          name="featured_slugs"
          defaultValue={settings.featured_categories.slugs.join(" ")}
          className="mt-1 w-full rounded-lg border border-border px-3 py-2"
        />
      </label>
      <label className="text-sm">
        درباره
        <textarea name="about" defaultValue={settings.pages.about} rows={4} className="mt-1 w-full rounded-lg border border-border px-3 py-2" />
      </label>
      <label className="text-sm">
        تماس
        <textarea name="contact" defaultValue={settings.pages.contact} rows={3} className="mt-1 w-full rounded-lg border border-border px-3 py-2" />
      </label>
      <label className="text-sm">
        قواعد استفاده
        <textarea name="terms" defaultValue={settings.pages.terms} rows={4} className="mt-1 w-full rounded-lg border border-border px-3 py-2" />
      </label>
      <label className="text-sm">
        سلب مسئولیت
        <textarea name="disclaimer" defaultValue={settings.pages.disclaimer} rows={4} className="mt-1 w-full rounded-lg border border-border px-3 py-2" />
      </label>
      <Button type="submit">ذخیره تنظیمات</Button>
    </form>
  );
}
