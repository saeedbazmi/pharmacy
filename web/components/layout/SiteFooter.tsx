import Link from "next/link";

export function SiteFooter({ disclaimer }: { disclaimer: string }) {
  return (
    <footer className="mt-auto border-t border-border bg-surface">
      <div className="mx-auto flex max-w-5xl flex-col gap-4 px-6 py-6">
        <nav aria-label="صفحات ثابت" className="flex flex-wrap gap-4 text-sm">
          <Link href="/about" className="text-muted hover:text-text">
            درباره
          </Link>
          <Link href="/contact" className="text-muted hover:text-text">
            تماس
          </Link>
          <Link href="/terms" className="text-muted hover:text-text">
            قواعد استفاده
          </Link>
        </nav>
        <p className="text-xs text-muted">{disclaimer}</p>
      </div>
    </footer>
  );
}
