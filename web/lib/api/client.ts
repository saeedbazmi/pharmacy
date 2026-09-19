import type {
  ApiErrorEnvelope,
  OfferQuery,
  Product,
  ProductListResponse,
  ProductSummary,
  SearchQuery,
} from "./types";
import type { SiteSettings } from "@/lib/site";
import { fallbackSite } from "@/lib/site";

/**
 * The only place the frontend talks to the backend. Components never call fetch
 * directly, so timeouts, error mapping and caching stay consistent.
 */

const BASE_URL = (process.env.API_BASE_URL ?? "http://localhost:8080").replace(
  /\/+$/,
  "",
);

const DEFAULT_TIMEOUT_MS = 10_000;

/**
 * An error the UI can show. `message` is already user-safe Persian text coming
 * from the backend envelope; `requestId` is kept so a user report can be traced
 * to server logs.
 */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly requestId?: string;

  constructor(args: {
    status: number;
    code: string;
    message: string;
    requestId?: string;
  }) {
    super(args.message);
    this.name = "ApiError";
    this.status = args.status;
    this.code = args.code;
    this.requestId = args.requestId;
  }

  /** True when the resource simply does not exist, so the UI can show 404. */
  get isNotFound(): boolean {
    return this.status === 404;
  }
}

const NETWORK_ERROR_MESSAGE =
  "ارتباط با سرور برقرار نشد. لطفاً دوباره تلاش کنید.";
const UNEXPECTED_ERROR_MESSAGE =
  "خطای غیرمنتظره‌ای رخ داد. لطفاً دوباره تلاش کنید.";

interface RequestOptions {
  /** Seconds before Next.js revalidates the cached response. */
  revalidate?: number;
  /** Cache tags, so a change in the panel can invalidate exactly this data. */
  tags?: string[];
  signal?: AbortSignal;
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { revalidate, tags, signal } = options;

  let response: Response;
  try {
    const headers: Record<string, string> = { Accept: "application/json" };
    const cookie = await userSessionCookie();
    if (cookie) {
      headers.Cookie = cookie;
    }
    response = await fetch(`${BASE_URL}${path}`, {
      headers,
      signal: signal ?? AbortSignal.timeout(DEFAULT_TIMEOUT_MS),
      next: revalidate === undefined && tags === undefined ? undefined : { revalidate, tags },
    });
  } catch {
    throw new ApiError({
      status: 0,
      code: "network_error",
      message: NETWORK_ERROR_MESSAGE,
    });
  }

  if (!response.ok) {
    throw await toApiError(response);
  }

  return (await response.json()) as T;
}

async function toApiError(response: Response): Promise<ApiError> {
  const requestId = response.headers.get("X-Request-Id") ?? undefined;
  try {
    const body = (await response.json()) as ApiErrorEnvelope;
    return new ApiError({
      status: response.status,
      code: body.error?.code ?? "unexpected_error",
      message: body.error?.message ?? UNEXPECTED_ERROR_MESSAGE,
      requestId: body.error?.request_id ?? requestId,
    });
  } catch {
    // A non-JSON body means something upstream broke; never show it to a user.
    return new ApiError({
      status: response.status,
      code: "unexpected_error",
      message: UNEXPECTED_ERROR_MESSAGE,
      requestId,
    });
  }
}

async function userSessionCookie(): Promise<string | undefined> {
  try {
    const { cookies } = await import("next/headers");
    const token = (await cookies()).get("user_session")?.value;
    if (token) {
      return `user_session=${token}`;
    }
  } catch {
    return undefined;
  }
  return undefined;
}

function offerQueryPath(filter?: OfferQuery): string {
  if (!filter) return "";
  const params = new URLSearchParams();
  if (filter.sort === "price_desc") params.set("sort", "price_desc");
  if (filter.inStock) params.set("in_stock", "1");
  if (filter.pharmacy) params.set("pharmacy", filter.pharmacy);
  if (filter.brand) params.set("brand", filter.brand);
  const encoded = params.toString();
  return encoded ? `?${encoded}` : "";
}

export const api = {
  /** Fetches a published product by slug, optionally filtered and sorted. */
  getProduct(
    slug: string,
    filter?: OfferQuery,
    options: RequestOptions = {},
  ): Promise<Product> {
    return request<Product>(
      `/api/v1/products/${encodeURIComponent(slug)}${offerQueryPath(filter)}`,
      {
        revalidate: 60,
        tags: [`product:${slug}`],
        ...options,
      },
    );
  },

  /** Compact catalogue rows for the home list. */
  async listProducts(limit = 48, options: RequestOptions = {}): Promise<ProductSummary[]> {
    const qs = limit === 48 ? "" : `?limit=${limit}`;
    const body = await request<ProductListResponse>(`/api/v1/products${qs}`, {
      revalidate: 60,
      tags: ["products"],
      ...options,
    });
    return body.products ?? [];
  },

  /** Compact rows for the compare table, keyed by slug. */
  async productsBySlugs(
    slugs: string[],
    options: RequestOptions = {},
  ): Promise<ProductSummary[]> {
    if (slugs.length === 0) return [];
    const body = await request<ProductListResponse>(
      `/api/v1/products?slugs=${encodeURIComponent(slugs.join(","))}`,
      {
        revalidate: 60,
        tags: slugs.map((slug) => `product:${slug}`),
        ...options,
      },
    );
    return body.products ?? [];
  },

  /** Full-text catalogue search. Cache key includes the normalised query. */
  searchProducts(query: SearchQuery, options: RequestOptions = {}): Promise<ProductListResponse> {
    const params = new URLSearchParams();
    params.set("q", query.q);
    if (query.sort !== "relevance") params.set("sort", query.sort);
    if (query.inStock) params.set("in_stock", "1");
    if (query.brand) params.set("brand", query.brand);
    if (query.category) params.set("category", query.category);
    if (query.page > 1) params.set("page", String(query.page));
    return request<ProductListResponse>(`/api/v1/products?${params.toString()}`, {
      revalidate: 30,
      tags: ["search"],
      ...options,
    });
  },

  /** Liveness of the backend, used by the development status page. */
  async health(): Promise<boolean> {
    try {
      await request<{ status: string }>("/healthz", { revalidate: 0 });
      return true;
    } catch {
      return false;
    }
  },

  /** Public site settings (banner, SEO, legal copy). */
  async siteSettings(options: RequestOptions = {}): Promise<SiteSettings> {
    try {
      return await request<SiteSettings>("/api/v1/site", {
        revalidate: 60,
        tags: ["site"],
        ...options,
      });
    } catch {
      return fallbackSite;
    }
  },

  /** Published products in one category. */
  async listByCategory(
    slug: string,
    options: RequestOptions = {},
  ): Promise<{ products: ProductSummary[]; category: { slug: string; name_fa: string } }> {
    return request(`/api/v1/products?category=${encodeURIComponent(slug)}`, {
      revalidate: 300,
      tags: ["products", `category:${slug}`],
      ...options,
    });
  },
};
