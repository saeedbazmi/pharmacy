import type { Metadata } from "next";
import Link from "next/link";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "@/components/ui/Table";
import { api } from "@/lib/api/client";
import { MAX_COMPARE, compareHref, parseCompareInput } from "@/lib/compare";
import { formatNumber, formatPrice } from "@/lib/format";

export const dynamic = "force-dynamic";

interface PageProps {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}

export async function generateMetadata({ searchParams }: PageProps): Promise<Metadata> {
  const { slugs } = parseCompareInput((await searchParams).c);
  return {
    title: "مقایسه کالاها",
    description: "کم‌ترین قیمت و تعداد داروخانه چند کالای هم‌رده را کنار هم ببینید.",
    alternates: {
      canonical: slugs.length ? compareHref(slugs) : "/compare",
    },
    robots: { index: false, follow: true },
    openGraph: {
      title: "مقایسه کالاها",
      locale: "fa_IR",
      type: "website",
    },
  };
}

export default async function ComparePage({ searchParams }: PageProps) {
  const { slugs, truncated } = parseCompareInput((await searchParams).c);
  const products = slugs.length ? await api.productsBySlugs(slugs) : [];
  const bySlug = new Map(products.map((product) => [product.slug, product]));
  const ordered = slugs
    .map((slug) => bySlug.get(slug))
    .filter((product): product is NonNullable<typeof product> => product !== undefined);

  return (
    <main className="mx-auto flex max-w-4xl flex-col gap-6 px-6 py-10">
      <h1 className="text-2xl font-bold">مقایسه کالاها</h1>
      <p className="text-muted">
        برای هر کالا کم‌ترین قیمت و تعداد داروخانه‌های دارنده نمایش داده می‌شود.
        سقف مقایسه {formatNumber(MAX_COMPARE)} کالاست.
      </p>

      {truncated ? (
        <p className="rounded-xl border border-warning bg-surface p-4 text-sm text-warning" role="status">
          بیش از {formatNumber(MAX_COMPARE)} کالا در لینک بود؛ فقط {formatNumber(MAX_COMPARE)} کالای اول مقایسه می‌شود.
        </p>
      ) : null}

      {slugs.length === 0 ? (
        <div className="flex flex-col gap-3 rounded-xl border border-border bg-surface p-6">
          <p className="text-muted">کالایی برای مقایسه انتخاب نشده است.</p>
          <Link href="/" className="text-primary hover:text-primary-dark">
            انتخاب از فهرست کالاها
          </Link>
        </div>
      ) : ordered.length === 0 ? (
        <div className="flex flex-col gap-3 rounded-xl border border-border bg-surface p-6">
          <p className="text-muted">هیچ‌کدام از کالاهای این لینک در فهرست عمومی نیست.</p>
          <Link href="/" className="text-primary hover:text-primary-dark">
            بازگشت به فهرست کالاها
          </Link>
        </div>
      ) : (
        <Table caption="مقایسه کم‌ترین قیمت و تعداد داروخانه کالاها">
          <TableHead>
            <TableRow>
              <TableHeaderCell>کالا</TableHeaderCell>
              <TableHeaderCell>برند</TableHeaderCell>
              <TableHeaderCell>کم‌ترین قیمت</TableHeaderCell>
              <TableHeaderCell>داروخانه‌ها</TableHeaderCell>
              <TableHeaderCell>
                <span className="sr-only">حذف</span>
              </TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {ordered.map((product) => {
              const remaining = slugs.filter((slug) => slug !== product.slug);
              return (
                <TableRow key={product.id}>
                  <TableCell>
                    <Link
                      href={`/product/${product.slug}`}
                      className="font-medium text-text hover:text-primary-dark"
                    >
                      {product.name_fa}
                    </Link>
                  </TableCell>
                  <TableCell>{product.brand_name || "—"}</TableCell>
                  <TableCell className="numeric">
                    {product.lowest_price_rial > 0
                      ? formatPrice(product.lowest_price_rial)
                      : "—"}
                  </TableCell>
                  <TableCell>
                    {formatNumber(product.offer_count)} داروخانه
                    {product.in_stock_count > 0
                      ? ` · ${formatNumber(product.in_stock_count)} موجود`
                      : ""}
                  </TableCell>
                  <TableCell>
                    <Link
                      href={compareHref(remaining)}
                      className="text-sm text-muted hover:text-danger"
                    >
                      حذف
                    </Link>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      )}
    </main>
  );
}
