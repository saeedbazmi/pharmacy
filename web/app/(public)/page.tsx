import type { Metadata } from "next";
import Link from "next/link";

import { ProductCard } from "@/components/product/ProductCard";
import { Badge } from "@/components/ui/Badge";
import { api } from "@/lib/api/client";
import type { ProductSummary } from "@/lib/api/types";

export const revalidate = 60;

export async function generateMetadata(): Promise<Metadata> {
  const site = await api.siteSettings();
  return {
    title: { absolute: site.seo.title },
    description: site.seo.description,
    alternates: { canonical: "/" },
    openGraph: {
      title: site.seo.title,
      description: site.seo.description,
      locale: "fa_IR",
      type: "website",
      url: "/",
    },
  };
}

export default async function HomePage() {
  const [site, products] = await Promise.all([
    api.siteSettings(),
    api.listProducts().catch((): ProductSummary[] => []),
  ]);

  return (
    <main className="mx-auto flex max-w-5xl flex-col gap-8 px-6 py-10">
      {site.banner.enabled && (site.banner.title || site.banner.body) ? (
        <aside className="rounded-xl border border-border bg-primary-soft p-5">
          {site.banner.title ? <h2 className="text-lg font-medium">{site.banner.title}</h2> : null}
          {site.banner.body ? <p className="mt-2 text-sm text-muted">{site.banner.body}</p> : null}
        </aside>
      ) : null}

      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <h1 className="text-3xl font-bold text-primary-dark">{site.seo.title}</h1>
            <Badge tone="info">فاز اول</Badge>
          </div>
        </div>
        <p className="text-muted">{site.seo.description}</p>
        <p className="text-sm text-muted">
          خرید و پرداخت در سایت داروخانه انجام می‌شود. قیمت و موجودی نهایی همان‌جا معتبر است.
        </p>
      </div>

      {site.featured_categories.slugs.length > 0 ? (
        <section className="flex flex-col gap-3">
          <h2 className="text-lg font-medium">دسته‌بندی‌های ویژه</h2>
          <ul className="flex flex-wrap gap-2">
            {site.featured_categories.slugs.map((slug) => (
              <li key={slug}>
                <Link
                  href={`/category/${slug}`}
                  className="inline-flex rounded-lg border border-border bg-surface px-3 py-1.5 text-sm hover:bg-primary-soft"
                >
                  {slug}
                </Link>
              </li>
            ))}
          </ul>
        </section>
      ) : null}

      {products.length === 0 ? (
        <p className="rounded-xl border border-border bg-surface p-6 text-muted">
          هنوز کالایی برای نمایش نیست. کمی بعد سر بزنید.
        </p>
      ) : (
        <section className="flex flex-col gap-4">
          <h2 className="text-lg font-medium">کالاهای موجود</h2>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {products.map((product) => (
              <ProductCard key={product.id} product={product} />
            ))}
          </div>
        </section>
      )}

      <p className="text-sm text-muted">
        چند کالا را انتخاب کنید و در صفحه{" "}
        <Link href="/compare" className="text-primary hover:text-primary-dark">
          مقایسه
        </Link>{" "}
        کنار هم ببینید.
      </p>
    </main>
  );
}
