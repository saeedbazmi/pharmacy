import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { CompareToggle } from "@/components/product/CompareToggle";
import { FavoriteButton } from "@/components/account/FavoriteButton";
import { AlertForm } from "@/components/account/AlertForm";
import { OfferDisclaimer } from "@/components/product/OfferDisclaimer";
import { OfferFilters } from "@/components/product/OfferFilters";
import { OfferTable } from "@/components/product/OfferTable";
import { ProductHeader } from "@/components/product/ProductHeader";
import { ProductJsonLd } from "@/components/product/ProductJsonLd";
import { api, ApiError } from "@/lib/api/client";
import { getMe, isFavorite } from "@/lib/account/server";
import {
  hasActiveOfferQuery,
  parseOfferQuery,
} from "@/lib/offer-query";

export const revalidate = 60;

interface PageProps {
  params: Promise<{ slug: string }>;
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { slug } = await params;
  try {
    const product = await api.getProduct(slug);
    return {
      title: product.name_fa,
      description: `مقایسه قیمت ${product.name_fa} در داروخانه‌های آنلاین. خرید در سایت داروخانه انجام می‌شود.`,
      alternates: { canonical: `/product/${slug}` },
      openGraph: {
        title: product.name_fa,
        description: `مقایسه قیمت ${product.name_fa}`,
        locale: "fa_IR",
        type: "website",
        url: `/product/${slug}`,
        images: product.image_url
          ? [{ url: product.image_url, alt: product.name_fa }]
          : undefined,
      },
    };
  } catch {
    return { title: "کالا یافت نشد" };
  }
}

export default async function ProductPage({ params, searchParams }: PageProps) {
  const { slug } = await params;
  const query = parseOfferQuery(await searchParams);
  const filtered = hasActiveOfferQuery(query);

  let product;
  try {
    product = await api.getProduct(slug);
  } catch (err) {
    if (err instanceof ApiError && err.isNotFound) {
      notFound();
    }
    throw err;
  }

  const tableProduct = filtered ? await api.getProduct(slug, query) : product;
  const pharmacies = uniquePharmacies(product.offers ?? []);
  const user = await getMe();
  const favorite = user ? await isFavorite(product.id) : false;

  return (
    <main className="mx-auto flex max-w-4xl flex-col gap-8 px-6 py-10">
      <ProductJsonLd product={product} />
      <ProductHeader product={product} />
      <div className="flex flex-wrap items-center gap-3">
        <CompareToggle slug={product.slug} name={product.name_fa} />
        <FavoriteButton
          productId={product.id}
          slug={product.slug}
          signedIn={Boolean(user)}
          isFavorite={favorite}
        />
      </div>
      <AlertForm productId={product.id} slug={product.slug} signedIn={Boolean(user)} />
      <section className="flex flex-col gap-3">
        <h2 className="text-lg font-medium">قیمت در داروخانه‌ها</h2>
        <OfferFilters
          slug={slug}
          query={query}
          pharmacies={pharmacies}
          brand={product.brand_name}
        />
        <OfferTable
          offers={tableProduct.offers ?? []}
          filtered={filtered}
          resetHref={`/product/${slug}`}
        />
        <OfferDisclaimer />
      </section>
      <p className="text-sm">
        <Link href="/" className="text-primary hover:text-primary-dark">
          بازگشت به فهرست کالاها
        </Link>
      </p>
    </main>
  );
}

function uniquePharmacies(
  offers: { pharmacy_slug: string; pharmacy_name: string }[],
) {
  const seen = new Map<string, string>();
  for (const offer of offers) {
    if (!seen.has(offer.pharmacy_slug)) {
      seen.set(offer.pharmacy_slug, offer.pharmacy_name);
    }
  }
  return [...seen.entries()].map(([slug, name]) => ({ slug, name }));
}
