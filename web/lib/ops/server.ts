import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import type { ApiErrorEnvelope } from "@/lib/api/types";
import { ApiError } from "@/lib/api/client";

const API_BASE = (process.env.API_BASE_URL ?? "http://localhost:8080").replace(/\/+$/, "");

export async function opsRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  const jar = await cookies();
  const token = jar.get("ops_session")?.value;
  if (!token) {
    redirect("/ops/login");
  }
  const headers = new Headers(init.headers);
  headers.set("Accept", "application/json");
  headers.set("Cookie", `ops_session=${token}`);
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers,
    cache: "no-store",
  });
  if (response.status === 401) {
    redirect("/ops/login");
  }
  if (response.status === 403) {
    redirect("/ops/login?error=forbidden");
  }
  if (!response.ok) {
    throw await opsError(response);
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

async function opsError(response: Response): Promise<ApiError> {
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
