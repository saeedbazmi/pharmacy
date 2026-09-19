"use client";

import { useEffect } from "react";

export default function CategoryError({
  reset,
}: {
  error: Error;
  reset: () => void;
}) {
  useEffect(() => {
    // keep the error boundary from leaking internals
  }, []);
  return (
    <main className="mx-auto flex max-w-5xl flex-col gap-4 px-6 py-10">
      <h1 className="text-2xl font-bold">بارگذاری دسته ممکن نشد</h1>
      <p className="text-muted">لطفاً دوباره تلاش کنید.</p>
      <button type="button" onClick={reset} className="self-start text-primary hover:text-primary-dark">
        تلاش مجدد
      </button>
    </main>
  );
}
