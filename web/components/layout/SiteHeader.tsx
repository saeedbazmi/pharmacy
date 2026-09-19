import Link from "next/link";

import { CompareNav } from "@/components/layout/CompareNav";
import { SearchBox } from "@/components/search/SearchBox";
import { getMe } from "@/lib/account/server";

export async function SiteHeader() {
  const user = await getMe();
  return (
    <header className="border-b border-border bg-surface">
      <div className="mx-auto flex max-w-5xl flex-col gap-3 px-6 py-3">
        <div className="flex items-center justify-between gap-4">
          <Link href="/" className="font-bold text-primary-dark">
            مقایسه قیمت دارو
          </Link>
          <nav aria-label="اصلی" className="flex items-center gap-4">
            <Link href="/" className="text-sm text-muted hover:text-text">
              کالاها
            </Link>
            <Link href="/search" className="text-sm text-muted hover:text-text">
              جست‌وجو
            </Link>
            <CompareNav />
            {user ? (
              <Link href="/account" className="text-sm text-muted hover:text-text">
                حساب
              </Link>
            ) : (
              <Link href="/login" className="text-sm text-muted hover:text-text">
                ورود
              </Link>
            )}
          </nav>
        </div>
        <SearchBox />
      </div>
    </header>
  );
}
