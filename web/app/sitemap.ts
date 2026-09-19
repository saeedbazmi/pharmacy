import type { MetadataRoute } from "next";

import { api } from "@/lib/api/client";

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000";

interface SitemapSlug {
  slug: string;
  updated_at: string;
}

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const base = (process.env.API_BASE_URL ?? "http://localhost:8080").replace(/\/+$/, "");
  let products: SitemapSlug[] = [];
  let categories: SitemapSlug[] = [];
  try {
    const body = await fetch(`${base}/api/v1/sitemap`, {
      next: { revalidate: 300, tags: ["sitemap"] },
    }).then((r) => r.json() as Promise<{ products: SitemapSlug[]; categories: SitemapSlug[] }>);
    products = body.products ?? [];
    categories = body.categories ?? [];
  } catch {
    const listed = await api.listProducts(48).catch(() => []);
    products = listed.map((p) => ({ slug: p.slug, updated_at: new Date().toISOString() }));
  }
  return [
    { url: siteUrl, changeFrequency: "daily", priority: 1 },
    { url: `${siteUrl}/search`, changeFrequency: "daily", priority: 0.8 },
    { url: `${siteUrl}/about`, changeFrequency: "monthly", priority: 0.3 },
    { url: `${siteUrl}/contact`, changeFrequency: "monthly", priority: 0.3 },
    { url: `${siteUrl}/terms`, changeFrequency: "monthly", priority: 0.3 },
    ...categories.map((c) => ({
      url: `${siteUrl}/category/${c.slug}`,
      lastModified: c.updated_at,
      changeFrequency: "daily" as const,
      priority: 0.5,
    })),
    ...products.map((p) => ({
      url: `${siteUrl}/product/${p.slug}`,
      lastModified: p.updated_at,
      changeFrequency: "daily" as const,
      priority: 0.6,
    })),
  ];
}
