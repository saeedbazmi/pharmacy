/** Wire types of the backend API. The single place response shapes are declared. */

export type ProductStatus = "published" | "hidden" | "needs_review";

export interface Offer {
  id: number;
  price_rial: number;
  in_stock: boolean;
  product_url: string;
  last_seen_at: string;
  pharmacy_name: string;
  pharmacy_slug: string;
  best_price: boolean;
  stale: boolean;
}

export type OfferSort = "price_asc" | "price_desc";

export interface OfferQuery {
  sort: OfferSort;
  inStock: boolean;
  pharmacy: string;
  brand: string;
}

export interface ProductSummary {
  id: number;
  slug: string;
  name_fa: string;
  image_url?: string;
  brand_name?: string;
  lowest_price_rial: number;
  offer_count: number;
  in_stock_count: number;
}

export interface ProductListResponse {
  products: ProductSummary[];
  page?: number;
  page_size?: number;
  total?: number;
  brands?: SearchFacet[];
  categories?: SearchFacet[];
}

export interface SearchFacet {
  slug: string;
  name: string;
}

export interface SearchQuery {
  q: string;
  sort: "relevance" | "price_asc" | "price_desc";
  inStock: boolean;
  brand: string;
  category: string;
  page: number;
}

export interface Product {
  id: number;
  slug: string;
  name_fa: string;
  name_en?: string;
  generic_name?: string;
  dosage_form?: string;
  strength?: string;
  image_url?: string;
  description?: string;
  brand_name?: string;
  category_name?: string;
  updated_at: string;
  offers: Offer[];
}

/** Shared error envelope: {"error": {"code", "message", "request_id"}}. */
export interface ApiErrorEnvelope {
  error: {
    code: string;
    message: string;
    request_id?: string;
  };
}

export interface PublicUser {
  id: number;
  phone_masked: string;
  role: string;
}

export interface FavoriteItem {
  product_id: number;
  slug: string;
  name_fa: string;
  image_url?: string;
  lowest_price_rial: number;
  offer_count: number;
  created_at: string;
}

export interface PriceAlertItem {
  id: number;
  product_id: number;
  kind: "price_drop" | "back_in_stock";
  target_price_rial?: number;
  baseline_price_rial: number;
  status: string;
  created_at: string;
  product_slug: string;
  product_name: string;
}

export interface SearchHistoryItem {
  id: number;
  at: string;
  query: string;
  results: number;
}
