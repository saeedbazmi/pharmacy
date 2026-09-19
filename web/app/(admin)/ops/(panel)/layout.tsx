import Link from "next/link";

import { opsRequest } from "@/lib/ops/server";
import type { OpsUser } from "@/lib/ops/types";

const NAV = [
  { href: "/ops", label: "خانه" },
  { href: "/ops/dashboard", label: "داشبورد", admin: true },
  { href: "/ops/health", label: "سلامت" },
  { href: "/ops/prices", label: "بازبینی قیمت" },
  { href: "/ops/clicks", label: "گزارش کلیک" },
  { href: "/ops/sources", label: "منابع" },
  { href: "/ops/matches", label: "صف تطبیق" },
  { href: "/ops/products", label: "کالاها" },
  { href: "/ops/categories", label: "دسته‌ها" },
  { href: "/ops/brands", label: "برندها" },
  { href: "/ops/users", label: "کاربران", admin: true },
  { href: "/ops/settings", label: "تنظیمات", admin: true },
  { href: "/ops/flags", label: "فلگ‌ها", admin: true },
  { href: "/ops/audit", label: "سابقه" },
];

export default async function OpsPanelLayout({ children }: { children: React.ReactNode }) {
  const me = await opsRequest<OpsUser>("/api/v1/ops/me");
  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border bg-surface">
        <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-3 px-6 py-3">
          <div className="flex items-center gap-6">
            <Link href="/ops" className="font-bold text-primary-dark">
              پنل داده
            </Link>
            <nav aria-label="پنل" className="flex flex-wrap gap-3 text-sm">
              {NAV.filter((item) => !item.admin || me.role === "super_admin").map((item) => (
                <Link key={item.href} href={item.href} className="text-muted hover:text-text">
                  {item.label}
                </Link>
              ))}
            </nav>
          </div>
          <div className="flex items-center gap-3 text-sm text-muted">
            <span>{me.username}</span>
            <form action="/ops/logout" method="post">
              <button type="submit" className="text-primary hover:text-primary-dark">
                خروج
              </button>
            </form>
          </div>
        </div>
      </header>
      {children}
    </div>
  );
}
