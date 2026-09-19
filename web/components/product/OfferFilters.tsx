import { Button } from "@/components/ui/Button";
import type { OfferQuery } from "@/lib/api/types";

interface PharmacyOption {
  slug: string;
  name: string;
}

export function OfferFilters({
  slug,
  query,
  pharmacies,
  brand,
}: {
  slug: string;
  query: OfferQuery;
  pharmacies: PharmacyOption[];
  brand?: string;
}) {
  return (
    <form
      method="get"
      action={`/product/${slug}`}
      className="flex flex-wrap items-end gap-4 rounded-xl border border-border bg-surface p-4"
    >
      <label className="flex flex-col gap-1 text-sm">
        <span className="text-muted">مرتب‌سازی</span>
        <select
          name="sort"
          defaultValue={query.sort}
          className="rounded-lg border border-border bg-background px-3 py-2"
        >
          <option value="price_asc">ارزان‌ترین</option>
          <option value="price_desc">گران‌ترین</option>
        </select>
      </label>

      <label className="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          name="in_stock"
          value="1"
          defaultChecked={query.inStock}
          className="size-4 accent-primary"
        />
        <span>فقط موجودها</span>
      </label>

      <label className="flex flex-col gap-1 text-sm">
        <span className="text-muted">داروخانه</span>
        <select
          name="pharmacy"
          defaultValue={query.pharmacy}
          className="rounded-lg border border-border bg-background px-3 py-2"
        >
          <option value="">همه داروخانه‌ها</option>
          {pharmacies.map((pharmacy) => (
            <option key={pharmacy.slug} value={pharmacy.slug}>
              {pharmacy.name}
            </option>
          ))}
        </select>
      </label>

      {brand ? (
        <label className="flex flex-col gap-1 text-sm">
          <span className="text-muted">برند</span>
          <select
            name="brand"
            defaultValue={query.brand}
            className="rounded-lg border border-border bg-background px-3 py-2"
          >
            <option value="">همه برندها</option>
            <option value={brand}>{brand}</option>
          </select>
        </label>
      ) : null}

      <Button type="submit" variant="secondary" size="sm">
        اعمال فیلتر
      </Button>
    </form>
  );
}
