"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/Button";
import type { AdminFlag, AdminPharmacy } from "@/lib/ops/types";

export function FlagsForm({ flags, pharmacies }: { flags: AdminFlag[]; pharmacies: AdminPharmacy[] }) {
  const router = useRouter();
  const [error, setError] = useState("");
  const global = flags.find((f) => f.key === "direct_purchase" && !f.pharmacy_id);

  async function setFlag(pharmacyId: number, enabled: boolean, confirmMsg: string) {
    if (!confirm(confirmMsg)) return;
    const response = await fetch("/api/v1/admin/flags", {
      method: "PUT",
      credentials: "include",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({ key: "direct_purchase", pharmacy_id: pharmacyId, enabled }),
    });
    if (!response.ok) {
      const err = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
      setError(err?.error?.message ?? "تغییر فلگ ناموفق بود.");
      return;
    }
    setError("");
    router.refresh();
  }

  return (
    <div className="flex flex-col gap-6">
      <p className="text-sm text-muted">
        فلگ «خرید مستقیم» فقط زیرساخت فاز دوم است. روشن کردنش در فاز اول هیچ سبد خرید، پرداخت یا
        سفارشی نمی‌سازد.
      </p>
      {error ? <p className="text-sm text-danger">{error}</p> : null}
      <section className="rounded-xl border border-border bg-surface p-4">
        <h2 className="text-lg font-medium">سراسری</h2>
        <p className="mt-1 text-sm text-muted">{global?.enabled ? "روشن" : "خاموش"}</p>
        <Button
          type="button"
          size="sm"
          className="mt-3"
          onClick={() =>
            setFlag(0, !global?.enabled, "فلگ سراسری خرید مستقیم تغییر کند؟ در فاز اول اثری روی سایت ندارد.")
          }
        >
          {global?.enabled ? "خاموش کردن" : "روشن کردن"}
        </Button>
      </section>
      <section>
        <h2 className="mb-3 text-lg font-medium">به‌تفکیک داروخانه</h2>
        <ul className="flex flex-col gap-3">
          {pharmacies.map((ph) => {
            const row = flags.find((f) => f.key === "direct_purchase" && f.pharmacy_id === ph.id);
            const on = Boolean(row?.enabled);
            return (
              <li key={ph.id} className="flex items-center justify-between rounded-xl border border-border bg-surface p-4">
                <div>
                  <div className="font-medium">{ph.name}</div>
                  <div className="text-sm text-muted">{on ? "روشن" : "خاموش (یا ارث از سراسری)"}</div>
                </div>
                <Button
                  type="button"
                  size="sm"
                  variant="secondary"
                  onClick={() => setFlag(ph.id, !on, `فلگ ${ph.name} تغییر کند؟`)}
                >
                  {on ? "خاموش" : "روشن"}
                </Button>
              </li>
            );
          })}
        </ul>
      </section>
    </div>
  );
}
