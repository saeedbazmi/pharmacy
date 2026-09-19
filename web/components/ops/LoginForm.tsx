"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/Button";

export function LoginForm() {
  const router = useRouter();
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError("");
    const data = new FormData(event.currentTarget);
    const response = await fetch("/api/v1/ops/login", {
      method: "POST",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      credentials: "include",
      body: JSON.stringify({
        username: String(data.get("username") ?? ""),
        password: String(data.get("password") ?? ""),
      }),
    });
    setPending(false);
    if (!response.ok) {
      try {
        const body = (await response.json()) as { error?: { message?: string } };
        setError(body.error?.message ?? "ورود ناموفق بود.");
      } catch {
        setError("ورود ناموفق بود.");
      }
      return;
    }
    router.replace("/ops");
    router.refresh();
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-4 rounded-xl border border-border bg-surface p-6">
      <label className="flex flex-col gap-1 text-sm">
        نام کاربری
        <input
          name="username"
          autoComplete="username"
          required
          className="rounded-lg border border-border px-3 py-2"
        />
      </label>
      <label className="flex flex-col gap-1 text-sm">
        گذرواژه
        <input
          name="password"
          type="password"
          autoComplete="current-password"
          required
          className="rounded-lg border border-border px-3 py-2"
        />
      </label>
      {error ? <p className="text-sm text-danger">{error}</p> : null}
      <Button type="submit" disabled={pending}>
        {pending ? "در حال ورود…" : "ورود"}
      </Button>
    </form>
  );
}
