import Link from "next/link";

import { opsRequest } from "@/lib/ops/server";
import type { OpsHealthSource, OpsMatch, OpsSource, OpsSuspiciousOffer } from "@/lib/ops/types";

export default async function OpsHomePage() {
  const [sources, matches, health, prices] = await Promise.all([
    opsRequest<{ sources: OpsSource[] }>("/api/v1/ops/sources"),
    opsRequest<{ matches: OpsMatch[]; total: number }>("/api/v1/ops/matches?page_size=5"),
    opsRequest<{ sources: OpsHealthSource[] }>("/api/v1/ops/health?days=7"),
    opsRequest<{ offers: OpsSuspiciousOffer[]; total: number }>("/api/v1/ops/prices?page_size=1"),
  ]);
  const overdue = health.sources.filter((s) => s.overdue || s.last_status === "failed").length;
  return (
    <main className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-8">
      <h1 className="text-2xl font-bold text-primary-dark">خانه پنل</h1>
      <p className="text-muted">
        یک منبع جدید بسازید، اتصال را تست کنید، بعد کالاهای صف تطبیق را تایید یا رد
        کنید. کلید API هیچ‌وقت خوانده نمی‌شود.
      </p>
      <div className="grid gap-4 sm:grid-cols-3">
        <Stat href="/ops/sources" label="منابع فعال" value={sources.sources.length} />
        <Stat href="/ops/matches" label="تطبیق در انتظار" value={matches.total} />
        <Stat href="/ops/health" label="منبع نیازمند توجه" value={overdue} />
        <Stat href="/ops/prices" label="قیمت مشکوک" value={prices.total} />
      </div>
    </main>
  );
}

function Stat({ href, label, value }: { href: string; label: string; value: number | string }) {
  return (
    <Link href={href} className="rounded-xl border border-border bg-surface p-4 hover:bg-primary-soft">
      <div className="text-sm text-muted">{label}</div>
      <div className="mt-1 text-2xl font-bold">{value}</div>
    </Link>
  );
}
