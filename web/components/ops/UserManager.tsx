"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/Button";
import type { AdminUser } from "@/lib/ops/types";

export function UserManager({ users }: { users: AdminUser[] }) {
  const router = useRouter();
  const [error, setError] = useState("");

  async function create(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const response = await fetch("/api/v1/admin/users", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({
        username: String(data.get("username") ?? ""),
        password: String(data.get("password") ?? ""),
        role: String(data.get("role") ?? ""),
      }),
    });
    if (!response.ok) {
      const err = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
      setError(err?.error?.message ?? "ساخت کاربر ناموفق بود.");
      return;
    }
    setError("");
    event.currentTarget.reset();
    router.refresh();
  }

  async function patch(id: number, body: Record<string, unknown>, confirmMsg: string) {
    if (!confirm(confirmMsg)) return;
    const response = await fetch(`/api/v1/admin/users/${id}`, {
      method: "PATCH",
      credentials: "include",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify(body),
    });
    if (!response.ok) {
      const err = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
      setError(err?.error?.message ?? "تغییر دسترسی ناموفق بود.");
      return;
    }
    setError("");
    router.refresh();
  }

  return (
    <div className="flex flex-col gap-8">
      {error ? <p className="text-sm text-danger">{error}</p> : null}
      <form onSubmit={create} className="flex flex-col gap-3 rounded-xl border border-border bg-surface p-4">
        <h2 className="text-lg font-medium">کاربر جدید</h2>
        <label className="text-sm">
          نام کاربری
          <input name="username" required minLength={3} className="mt-1 w-full rounded-lg border border-border px-3 py-2" />
        </label>
        <label className="text-sm">
          گذرواژه
          <input name="password" type="password" required minLength={10} className="mt-1 w-full rounded-lg border border-border px-3 py-2" />
        </label>
        <label className="text-sm">
          نقش
          <select name="role" defaultValue="data_ops" className="mt-1 w-full rounded-lg border border-border px-3 py-2">
            <option value="data_ops">تیم داده</option>
            <option value="super_admin">مدیر کل</option>
          </select>
        </label>
        <Button type="submit">ساخت</Button>
      </form>
      <ul className="flex flex-col gap-3">
        {users.map((user) => (
          <li key={user.id} className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-border bg-surface p-4">
            <div>
              <div className="font-medium">{user.username}</div>
              <div className="text-sm text-muted">
                {user.role === "super_admin" ? "مدیر کل" : "تیم داده"}
                {user.is_active ? "" : " · غیرفعال"}
              </div>
            </div>
            <div className="flex flex-wrap gap-2">
              {user.role === "super_admin" ? (
                <Button
                  type="button"
                  size="sm"
                  variant="secondary"
                  onClick={() =>
                    patch(user.id, { role: "data_ops", is_active: user.is_active }, "نقش این کاربر به تیم داده تغییر کند؟")
                  }
                >
                  تنزل نقش
                </Button>
              ) : (
                <Button
                  type="button"
                  size="sm"
                  variant="secondary"
                  onClick={() =>
                    patch(user.id, { role: "super_admin", is_active: user.is_active }, "این کاربر مدیر کل شود؟")
                  }
                >
                  ارتقا به مدیر کل
                </Button>
              )}
              <Button
                type="button"
                size="sm"
                variant="secondary"
                onClick={() =>
                  patch(
                    user.id,
                    { role: user.role, is_active: !user.is_active },
                    user.is_active ? "این حساب غیرفعال شود؟" : "این حساب فعال شود؟",
                  )
                }
              >
                {user.is_active ? "غیرفعال" : "فعال"}
              </Button>
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
