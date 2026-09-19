import { cookies } from "next/headers";

import type {
  ApiErrorEnvelope,
  FavoriteItem,
  PriceAlertItem,
  PublicUser,
  SearchHistoryItem,
} from "@/lib/api/types";
import { ApiError } from "@/lib/api/client";

const API_BASE = (process.env.API_BASE_URL ?? "http://localhost:8080").replace(/\/+$/, "");

export function safeNextPath(raw: string | null | undefined, fallback = "/account"): string {
  if (!raw) return fallback;
  if (!raw.startsWith("/") || raw.startsWith("//")) return fallback;
  if (raw.startsWith("/ops") || raw.startsWith("/admin")) return fallback;
  return raw;
}

export async function getMe(): Promise<PublicUser | null> {
  const token = (await cookies()).get("user_session")?.value;
  if (!token) return null;
  const response = await fetch(`${API_BASE}/api/v1/auth/me`, {
    headers: { Accept: "application/json", Cookie: `user_session=${token}` },
    cache: "no-store",
  });
  if (response.status === 401) return null;
  if (!response.ok) return null;
  return (await response.json()) as PublicUser;
}

export async function isFavorite(productId: number): Promise<boolean> {
  const token = (await cookies()).get("user_session")?.value;
  if (!token) return false;
  const response = await fetch(`${API_BASE}/api/v1/account/favorites/${productId}`, {
    headers: { Accept: "application/json", Cookie: `user_session=${token}` },
    cache: "no-store",
  });
  if (!response.ok) return false;
  const body = (await response.json()) as { favorite?: boolean };
  return Boolean(body.favorite);
}

export async function listFavorites(): Promise<FavoriteItem[]> {
  const body = await accountGet<{ favorites: FavoriteItem[] }>("/api/v1/account/favorites");
  return body?.favorites ?? [];
}

export async function listAlerts(): Promise<PriceAlertItem[]> {
  const body = await accountGet<{ alerts: PriceAlertItem[] }>("/api/v1/account/alerts");
  return body?.alerts ?? [];
}

export async function listSearches(): Promise<SearchHistoryItem[]> {
  const body = await accountGet<{ searches: SearchHistoryItem[] }>("/api/v1/account/searches");
  return body?.searches ?? [];
}

export async function accountGet<T>(path: string): Promise<T | null> {
  const token = (await cookies()).get("user_session")?.value;
  if (!token) return null;
  const response = await fetch(`${API_BASE}${path}`, {
    headers: { Accept: "application/json", Cookie: `user_session=${token}` },
    cache: "no-store",
  });
  if (response.status === 401) return null;
  if (!response.ok) {
    throw await readError(response);
  }
  return (await response.json()) as T;
}

export async function accountMutate(path: string, init: RequestInit): Promise<Response> {
  const token = (await cookies()).get("user_session")?.value;
  const headers = new Headers(init.headers);
  headers.set("Accept", "application/json");
  if (token) {
    headers.set("Cookie", `user_session=${token}`);
  }
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  return fetch(`${API_BASE}${path}`, { ...init, headers, cache: "no-store" });
}

export async function readError(response: Response): Promise<ApiError> {
  try {
    const body = (await response.json()) as ApiErrorEnvelope;
    return new ApiError({
      status: response.status,
      code: body.error?.code ?? "unexpected_error",
      message: body.error?.message ?? "خطای غیرمنتظره",
    });
  } catch {
    return new ApiError({
      status: response.status,
      code: "unexpected_error",
      message: "خطای غیرمنتظره",
    });
  }
}

export { API_BASE };
