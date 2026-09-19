import type { OfferQuery, OfferSort } from "@/lib/api/types";

type SearchValue = string | string[] | undefined;

function first(value: SearchValue): string {
  if (Array.isArray(value)) return value[0] ?? "";
  return value ?? "";
}

/** Maps URL search params to the catalog offer filter. Invalid values become defaults. */
export function parseOfferQuery(
  params: Record<string, SearchValue>,
): OfferQuery {
  const sortRaw = first(params.sort);
  const sort: OfferSort = sortRaw === "price_desc" ? "price_desc" : "price_asc";
  const inRaw = first(params.in_stock).toLowerCase();
  return {
    sort,
    inStock: inRaw === "1" || inRaw === "true" || inRaw === "yes",
    pharmacy: first(params.pharmacy).trim().toLowerCase(),
    brand: first(params.brand).trim(),
  };
}

export function hasActiveOfferQuery(query: OfferQuery): boolean {
  return (
    query.sort !== "price_asc" ||
    query.inStock ||
    query.pharmacy !== "" ||
    query.brand !== ""
  );
}

export function offerQueryString(query: OfferQuery): string {
  const params = new URLSearchParams();
  if (query.sort === "price_desc") params.set("sort", "price_desc");
  if (query.inStock) params.set("in_stock", "1");
  if (query.pharmacy) params.set("pharmacy", query.pharmacy);
  if (query.brand) params.set("brand", query.brand);
  const encoded = params.toString();
  return encoded ? `?${encoded}` : "";
}
