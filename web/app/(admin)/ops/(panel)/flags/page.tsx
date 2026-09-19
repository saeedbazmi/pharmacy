import { FlagsForm } from "@/components/ops/FlagsForm";
import { opsRequest } from "@/lib/ops/server";
import type { AdminFlag, AdminPharmacy, OpsUser } from "@/lib/ops/types";

export default async function FlagsPage() {
  const me = await opsRequest<OpsUser>("/api/v1/ops/me");
  if (me.role !== "super_admin") {
    return (
      <main className="mx-auto max-w-6xl px-6 py-8">
        <h1 className="text-2xl font-bold">فلگ‌ها</h1>
        <p className="mt-3 text-muted">فقط مدیر کل به این بخش دسترسی دارد.</p>
      </main>
    );
  }
  const body = await opsRequest<{ flags: AdminFlag[]; pharmacies: AdminPharmacy[] }>(
    "/api/v1/admin/flags",
  );
  return (
    <main className="mx-auto flex max-w-3xl flex-col gap-6 px-6 py-8">
      <h1 className="text-2xl font-bold text-primary-dark">فلگ قابلیت‌ها</h1>
      <FlagsForm flags={body.flags ?? []} pharmacies={body.pharmacies ?? []} />
    </main>
  );
}
