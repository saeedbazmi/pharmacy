import Link from "next/link";

import { CompareToggle } from "@/components/product/CompareToggle";
import { ProductImage } from "@/components/product/ProductImage";
import { Badge } from "@/components/ui/Badge";
import { formatNumber, formatPrice } from "@/lib/format";
import type { ProductSummary } from "@/lib/api/types";

export function SearchResultCard({ product }: { product: ProductSummary }) {
  const inStock = product.in_stock_count > 0;
  return (
    <article className="flex flex-col gap-3 rounded-xl border border-border bg-surface p-4">
      <div className="flex size-28 items-center justify-center overflow-hidden rounded-lg border border-border bg-background">
        <ProductImage src={product.image_url} alt={product.name_fa} width={112} height={112} />
      </div>
      <h2 className="text-base font-medium">
        <Link href={`/product/${product.slug}`} className="text-text hover:text-primary-dark">
          {product.name_fa}
        </Link>
      </h2>
      {product.brand_name ? <Badge>{product.brand_name}</Badge> : null}
      <p className="numeric text-sm">
        {product.lowest_price_rial > 0 ? `از ${formatPrice(product.lowest_price_rial)}` : "بدون قیمت"}
      </p>
      <p className="text-sm text-muted">{formatNumber(product.offer_count)} داروخانه</p>
      <p className="text-sm">
        {inStock ? (
          <span className="text-success">موجود در داروخانه</span>
        ) : (
          <span className="text-danger">ناموجود</span>
        )}
      </p>
      <CompareToggle slug={product.slug} name={product.name_fa} />
    </article>
  );
}
