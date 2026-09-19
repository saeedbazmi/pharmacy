import type { Product } from "@/lib/api/types";

/** Server-only JSON-LD for the public product page. */
export function ProductJsonLd({ product }: { product: Product }) {
  const prices = (product.offers ?? [])
    .map((offer) => offer.price_rial)
    .filter((price) => price > 0);
  const lowPrice = prices.length ? Math.min(...prices) : undefined;
  const highPrice = prices.length ? Math.max(...prices) : undefined;

  const payload = {
    "@context": "https://schema.org",
    "@type": "Product",
    name: product.name_fa,
    image: product.image_url || undefined,
    brand: product.brand_name
      ? { "@type": "Brand", name: product.brand_name }
      : undefined,
    offers: {
      "@type": "AggregateOffer",
      priceCurrency: "IRR",
      lowPrice,
      highPrice,
      offerCount: product.offers?.length ?? 0,
      availability: product.offers?.some((offer) => offer.in_stock)
        ? "https://schema.org/InStock"
        : "https://schema.org/OutOfStock",
    },
  };

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(payload) }}
    />
  );
}
