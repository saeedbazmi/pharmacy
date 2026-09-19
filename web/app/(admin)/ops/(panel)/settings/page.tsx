import { SettingsForm } from "@/components/ops/SettingsForm";
import { opsRequest } from "@/lib/ops/server";
import type { OpsUser } from "@/lib/ops/types";
import type { SiteSettings } from "@/lib/site";

export default async function SettingsPage() {
  const me = await opsRequest<OpsUser>("/api/v1/ops/me");
  if (me.role !== "super_admin") {
    return (
      <main className="mx-auto max-w-6xl px-6 py-8">
        <h1 className="text-2xl font-bold">تنظیمات سایت</h1>
        <p className="mt-3 text-muted">فقط مدیر کل به این بخش دسترسی دارد.</p>
      </main>
    );
  }
  const settings = await opsRequest<SiteSettings>("/api/v1/admin/settings");
  return (
    <main className="mx-auto flex max-w-3xl flex-col gap-6 px-6 py-8">
      <h1 className="text-2xl font-bold text-primary-dark">تنظیمات سایت</h1>
      <p className="text-muted">پس از ذخیره، صفحات عمومی بدون استقرار مجدد تازه می‌شوند.</p>
      <SettingsForm settings={settings} />
    </main>
  );
}
