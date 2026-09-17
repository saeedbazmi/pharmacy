import type { ApiErrorEnvelope, Product } from "./types";

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
    response = await fetch(`${BASE_URL}${path}`, {
      headers: { Accept: "application/json" },
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

export const api = {
  /** Fetches a published product by slug. */
  getProduct(slug: string, options: RequestOptions = {}): Promise<Product> {
    return request<Product>(`/api/v1/products/${encodeURIComponent(slug)}`, {
      revalidate: 60,
      tags: [`product:${slug}`],
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
};
