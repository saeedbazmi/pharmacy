import type { Metadata } from "next";

import { LoginForm } from "@/components/ops/LoginForm";

export const metadata: Metadata = {
  title: "ورود پنل داده",
  robots: { index: false, follow: false },
};

export default function OpsLoginPage() {
  return (
    <main className="mx-auto flex min-h-screen max-w-md flex-col justify-center gap-6 px-6">
      <div>
        <h1 className="text-2xl font-bold text-primary-dark">ورود پنل داده</h1>
        <p className="mt-2 text-sm text-muted">
          این بخش فقط برای تیم محتواست و در نتایج جست‌وجو دیده نمی‌شود.
        </p>
      </div>
      <LoginForm />
    </main>
  );
}
