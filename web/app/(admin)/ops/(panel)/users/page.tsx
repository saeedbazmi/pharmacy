import { UserManager } from "@/components/ops/UserManager";
import { opsRequest } from "@/lib/ops/server";
import type { AdminUser, OpsUser } from "@/lib/ops/types";

export default async function UsersPage() {
  const me = await opsRequest<OpsUser>("/api/v1/ops/me");
  if (me.role !== "super_admin") {
    return (
      <main className="mx-auto max-w-6xl px-6 py-8">
        <h1 className="text-2xl font-bold">کاربران داخلی</h1>
        <p className="mt-3 text-muted">فقط مدیر کل به این بخش دسترسی دارد.</p>
      </main>
    );
  }
  const body = await opsRequest<{ users: AdminUser[] }>("/api/v1/admin/users");
  return (
    <main className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-8">
      <h1 className="text-2xl font-bold text-primary-dark">کاربران داخلی</h1>
      <p className="text-muted">آخرین مدیر کل را نمی‌توان غیرفعال یا تنزل داد.</p>
      <UserManager users={body.users ?? []} />
    </main>
  );
}
