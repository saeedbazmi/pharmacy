import type { Metadata } from "next";
import Link from "next/link";

import { SearchBox } from "@/components/search/SearchBox";
import { SearchFilters } from "@/components/search/SearchFilters";
import { SearchPagination } from "@/components/search/SearchPagination";
import { SearchResultCard } from "@/components/search/SearchResultCard";
import { api, ApiError } from "@/lib/api/client";
import { formatNumber } from "@/lib/format";
import { normalizeQuery } from "@/lib/persian";
import type { SearchQuery } from "@/lib/api/types";

export const dynamic = "force-dynamic";

interface PageProps {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}

export async function generateMetadata({ searchParams }: PageProps): Promise<Metadata> {
  const q = first((await searchParams).q);
  return {
    title: q ? `جست‌وجو: ${q}` : "جست‌وجو",
    description: "جست‌وجوی دارو و محصولات سلامت در داروخانه‌های آنلاین.",
    alternates: { canonical: q ? `/search?q=${encodeURIComponent(q)}` : "/search" },
    openGraph: {
      title: q ? `جست‌وجو: ${q}` : "جست‌وجوی دارو",
      description: "جست‌وجوی دارو و محصولات سلامت در داروخانه‌های آنلاین.",
      locale: "fa_IR",
      type: "website",
      url: q ? `/search?q=${encodeURIComponent(q)}` : "/search",
    },
  };
}

export default async function SearchPage({ searchParams }: PageProps) {
  const raw = await searchParams;
  const query = parseSearchQuery(raw);
  if (!query.q) {
    return (
      <main className="mx-auto flex max-w-4xl flex-col gap-6 px-6 py-10">
        <h1 className="text-2xl font-bold">جست‌وجوی دارو</h1>
        <SearchBox />
        <p className="text-muted">نام دارو، برند یا ماده مؤثره را بنویسید.</p>
      </main>
    );
  }

  let result;
  try {
    result = await api.searchProducts(query);
  } catch (err) {
    if (err instanceof ApiError && err.status === 400) {
      return (
        <main className="mx-auto flex max-w-4xl flex-col gap-6 px-6 py-10">
          <h1 className="text-2xl font-bold">جست‌وجوی دارو</h1>
          <SearchBox defaultQuery={query.q} />
          <p className="text-muted">{err.message}</p>
        </main>
      );
    }
    throw err;
  }

  const total = result.total ?? 0;
  const page = result.page ?? query.page;
  const pageSize = result.page_size ?? 24;
  const hrefFor = (pageNum: number) => searchHref({ ...query, page: pageNum });

  return (
    <main className="mx-auto flex max-w-5xl flex-col gap-6 px-6 py-10">
      <h1 className="text-2xl font-bold">نتایج جست‌وجو</h1>
      <SearchBox defaultQuery={query.q} />
      <SearchFilters
        q={query.q}
        sort={query.sort}
        inStock={query.inStock}
        brand={query.brand}
        category={query.category}
        brands={result.brands ?? []}
        categories={result.categories ?? []}
      />
      <p className="text-sm text-muted">{formatNumber(total)} کالا</p>
      {result.products.length === 0 ? (
        <div className="flex flex-col gap-3 rounded-xl border border-border bg-surface p-6">
          <p>نتیجه‌ای برای «{query.q}» پیدا نشد.</p>
          <p className="text-sm text-muted">
            املا را بررسی کنید یا فاصله و ی/ک عربی را یکسان بنویسید. می‌توانید از فهرست کالاها هم
            شروع کنید.
          </p>
          <Link href="/" className="text-primary hover:text-primary-dark">
            مرور کالاها
          </Link>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {result.products.map((product) => (
            <SearchResultCard key={product.id} product={product} />
          ))}
        </div>
      )}
      <SearchPagination
        hrefFor={hrefFor}
        page={page}
        pageSize={pageSize}
        total={total}
      />
    </main>
  );
}

function first(value: string | string[] | undefined): string {
  if (Array.isArray(value)) return value[0] ?? "";
  return value ?? "";
}

function parseSearchQuery(raw: Record<string, string | string[] | undefined>): SearchQuery {
  const sortRaw = first(raw.sort);
  const sort: SearchQuery["sort"] =
    sortRaw === "price_asc" || sortRaw === "price_desc" ? sortRaw : "relevance";
  const inRaw = first(raw.in_stock).toLowerCase();
  const page = Number.parseInt(first(raw.page), 10);
  return {
    q: normalizeQuery(first(raw.q)),
    sort,
    inStock: inRaw === "1" || inRaw === "true" || inRaw === "yes",
    brand: first(raw.brand).trim().toLowerCase(),
    category: first(raw.category).trim().toLowerCase(),
    page: Number.isFinite(page) && page > 0 ? page : 1,
  };
}

function searchHref(q: SearchQuery): string {
  const params = new URLSearchParams();
  params.set("q", q.q);
  if (q.sort !== "relevance") params.set("sort", q.sort);
  if (q.inStock) params.set("in_stock", "1");
  if (q.brand) params.set("brand", q.brand);
  if (q.category) params.set("category", q.category);
  if (q.page > 1) params.set("page", String(q.page));
  return `/search?${params.toString()}`;
}
