import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";

import { OTPLoginForm } from "@/components/account/OTPLoginForm";
import { getMe, safeNextPath } from "@/lib/account/server";

export const metadata: Metadata = {
  title: "ورود",
  robots: { index: false, follow: false },
};

interface PageProps {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}

function one(v: string | string[] | undefined): string {
  if (Array.isArray(v)) return v[0] ?? "";
  return v ?? "";
}

export default async function LoginPage({ searchParams }: PageProps) {
  const me = await getMe();
  const params = await searchParams;
  const nextPath = safeNextPath(one(params.next), "/account");
  if (me) {
    redirect(nextPath);
  }

  return (
    <main className="mx-auto flex min-h-[70vh] max-w-md flex-col justify-center gap-6 px-6 py-10">
      <div>
        <h1 className="text-2xl font-bold text-primary-dark">ورود با شماره موبایل</h1>
        <p className="mt-2 text-sm text-muted">
          ورود اختیاری است. جست‌وجو، مقایسه و خرید بدون حساب هم کار می‌کند. حساب فقط برای
          علاقه‌مندی و هشدار قیمت است.
        </p>
      </div>
      <OTPLoginForm
        nextPath={nextPath}
        intent={one(params.intent)}
        productId={one(params.product_id)}
      />
      <p className="text-sm">
        <Link href="/" className="text-primary hover:text-primary-dark">
          ادامه بدون ورود
        </Link>
      </p>
    </main>
  );
}
