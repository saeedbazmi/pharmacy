"use client";

export default function AccountError({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <main className="mx-auto flex max-w-lg flex-col gap-4 px-6 py-16">
      <h1 className="text-xl font-bold">حساب بارگذاری نشد</h1>
      <p className="text-muted">لطفاً دوباره تلاش کنید. جست‌وجو و خرید بدون ورود همچنان در دسترس است.</p>
      <button
        type="button"
        onClick={reset}
        className="self-start rounded-lg bg-primary px-4 py-2 text-white hover:bg-primary-dark"
      >
        تلاش مجدد
      </button>
    </main>
  );
}
