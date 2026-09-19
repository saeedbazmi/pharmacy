import type { Metadata } from "next";

import { api } from "@/lib/api/client";
import type { SitePages } from "@/lib/site";

export const revalidate = 60;

type PageKey = keyof SitePages;

const PAGES: Record<
  PageKey,
  { path: string; title: string; heading: string }
> = {
  about: { path: "/about", title: "درباره", heading: "درباره این وب‌سایت" },
  contact: { path: "/contact", title: "تماس", heading: "تماس با ما" },
  terms: { path: "/terms", title: "قواعد استفاده", heading: "قواعد استفاده" },
  disclaimer: { path: "/about", title: "سلب مسئولیت", heading: "سلب مسئولیت" },
};

export function legalMetadata(key: Exclude<PageKey, "disclaimer">): Promise<Metadata> {
  return buildMetadata(key);
}

async function buildMetadata(key: Exclude<PageKey, "disclaimer">): Promise<Metadata> {
  const site = await api.siteSettings();
  const meta = PAGES[key];
  return {
    title: meta.title,
    description: site.pages[key].slice(0, 160),
    alternates: { canonical: meta.path },
    openGraph: {
      title: meta.title,
      description: site.pages[key].slice(0, 160),
      locale: "fa_IR",
      type: "website",
      url: meta.path,
    },
  };
}

export async function LegalPage({ page }: { page: Exclude<PageKey, "disclaimer"> }) {
  const site = await api.siteSettings();
  const meta = PAGES[page];
  return (
    <main className="mx-auto flex max-w-3xl flex-col gap-6 px-6 py-10">
      <h1 className="text-2xl font-bold text-primary-dark">{meta.heading}</h1>
      <p className="whitespace-pre-wrap leading-8 text-text">{site.pages[page]}</p>
      <p className="text-sm text-muted">{site.pages.disclaimer}</p>
      <p className="text-sm text-muted">
        خرید در سایت داروخانه انجام می‌شود و قیمت نهایی همان‌جا معتبر است. این پلتفرم
        سبد خرید، پرداخت یا سفارش ندارد.
      </p>
    </main>
  );
}
