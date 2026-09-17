/** Wire types of the backend API. The single place response shapes are declared. */

export type ProductStatus = "published" | "hidden" | "needs_review";

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
}

/** Shared error envelope: {"error": {"code", "message", "request_id"}}. */
export interface ApiErrorEnvelope {
  error: {
    code: string;
    message: string;
    request_id?: string;
  };
}
