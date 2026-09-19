import Link from "next/link";

export default function ProductNotFound() {
  return (
    <main className="mx-auto flex max-w-lg flex-col gap-4 px-6 py-16">
      <h1 className="text-xl font-bold">کالا یافت نشد</h1>
      <p className="text-muted">این کالا در فهرست ما نیست یا از نمایش عمومی خارج شده است.</p>
      <Link href="/" className="text-primary hover:text-primary-dark">
        بازگشت به صفحه اصلی
      </Link>
    </main>
  );
}

