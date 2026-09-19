import Link from "next/link";

import { formatNumber } from "@/lib/format";

export function SearchPagination({
  hrefFor,
  page,
  pageSize,
  total,
}: {
  hrefFor: (page: number) => string;
  page: number;
  pageSize: number;
  total: number;
}) {
  const last = Math.max(1, Math.ceil(total / pageSize));
  if (last <= 1) return null;
  const pages: number[] = [];
  const from = Math.max(1, page - 2);
  const to = Math.min(last, page + 2);
  for (let n = from; n <= to; n++) pages.push(n);

  return (
    <nav aria-label="صفحه‌بندی نتایج" className="flex flex-wrap items-center gap-2">
      {page > 1 ? (
        <Link href={hrefFor(page - 1)} className="rounded-lg border border-border px-3 py-1.5 text-sm">
          قبلی
        </Link>
      ) : null}
      {pages.map((n) => (
        <Link
          key={n}
          href={hrefFor(n)}
          aria-current={n === page ? "page" : undefined}
          className={`rounded-lg border px-3 py-1.5 text-sm ${
            n === page
              ? "border-primary bg-primary-soft text-primary-dark"
              : "border-border text-text"
          }`}
        >
          {formatNumber(n)}
        </Link>
      ))}
      {page < last ? (
        <Link href={hrefFor(page + 1)} className="rounded-lg border border-border px-3 py-1.5 text-sm">
          بعدی
        </Link>
      ) : null}
    </nav>
  );
}
