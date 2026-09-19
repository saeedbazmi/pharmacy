import { Button } from "@/components/ui/Button";

export type SearchSort = "relevance" | "price_asc" | "price_desc";

export function SearchFilters({
  q,
  sort,
  inStock,
  brand,
  category,
  brands,
  categories,
}: {
  q: string;
  sort: SearchSort;
  inStock: boolean;
  brand: string;
  category: string;
  brands: { slug: string; name: string }[];
  categories: { slug: string; name: string }[];
}) {
  return (
    <form
      method="get"
      action="/search"
      className="flex flex-wrap items-end gap-4 rounded-xl border border-border bg-surface p-4"
    >
      <input type="hidden" name="q" value={q} />
      <label className="flex flex-col gap-1 text-sm">
        <span className="text-muted">مرتب‌سازی</span>
        <select
          name="sort"
          defaultValue={sort}
          className="rounded-lg border border-border bg-background px-3 py-2"
        >
          <option value="relevance">مرتبط‌ترین</option>
          <option value="price_asc">ارزان‌ترین</option>
          <option value="price_desc">گران‌ترین</option>
        </select>
      </label>
      <label className="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          name="in_stock"
          value="1"
          defaultChecked={inStock}
          className="size-4 accent-primary"
        />
        <span>فقط موجودها</span>
      </label>
      <label className="flex flex-col gap-1 text-sm">
        <span className="text-muted">برند</span>
        <select
          name="brand"
          defaultValue={brand}
          className="rounded-lg border border-border bg-background px-3 py-2"
        >
          <option value="">همه برندها</option>
          {brands.map((item) => (
            <option key={item.slug} value={item.slug}>
              {item.name}
            </option>
          ))}
        </select>
      </label>
      <label className="flex flex-col gap-1 text-sm">
        <span className="text-muted">دسته‌بندی</span>
        <select
          name="category"
          defaultValue={category}
          className="rounded-lg border border-border bg-background px-3 py-2"
        >
          <option value="">همه دسته‌ها</option>
          {categories.map((item) => (
            <option key={item.slug} value={item.slug}>
              {item.name}
            </option>
          ))}
        </select>
      </label>
      <Button type="submit" variant="secondary" size="sm">
        اعمال فیلتر
      </Button>
    </form>
  );
}
