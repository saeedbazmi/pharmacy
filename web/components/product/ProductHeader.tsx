import { Badge } from "@/components/ui/Badge";
import { ProductImage } from "@/components/product/ProductImage";
import type { Product } from "@/lib/api/types";

export function ProductHeader({ product }: { product: Product }) {
  return (
    <header className="flex flex-col gap-6 md:flex-row md:items-start">
      <div className="flex size-48 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-border bg-surface">
        <ProductImage
          src={product.image_url}
          alt={product.name_fa}
          width={192}
          height={192}
          priority
        />
      </div>
      <div className="flex min-w-0 flex-col gap-3">
        <h1 className="text-2xl font-bold text-text">{product.name_fa}</h1>
        {product.name_en ? (
          <p className="text-muted" lang="en" dir="ltr">
            {product.name_en}
          </p>
        ) : null}
        <dl className="flex flex-wrap gap-3 text-sm">
          {product.brand_name ? (
            <div>
              <dt className="text-muted">برند</dt>
              <dd>{product.brand_name}</dd>
            </div>
          ) : null}
          {product.generic_name ? (
            <div>
              <dt className="text-muted">ماده مؤثره</dt>
              <dd>{product.generic_name}</dd>
            </div>
          ) : null}
          {product.dosage_form ? (
            <div>
              <dt className="text-muted">شکل دارویی</dt>
              <dd>{product.dosage_form}</dd>
            </div>
          ) : null}
          {product.strength ? (
            <div>
              <dt className="text-muted">دوز</dt>
              <dd>{product.strength}</dd>
            </div>
          ) : null}
        </dl>
        {product.category_name ? <Badge>{product.category_name}</Badge> : null}
      </div>
    </header>
  );
}
