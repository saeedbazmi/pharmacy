"use server";

import { cookies, headers } from "next/headers";
import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { API_BASE, accountMutate, readError, safeNextPath } from "@/lib/account/server";
import { toLatinDigits } from "@/lib/persian";

export type OTPState = {
  sent?: boolean;
  phone?: string;
  phoneMasked?: string;
  resendAfter?: number;
  error?: string;
};

function clientIP(h: Headers): string {
  const xff = h.get("x-forwarded-for");
  if (xff) return xff.split(",")[0]?.trim() ?? "";
  return h.get("x-real-ip") ?? "";
}

export async function requestOTP(_prev: OTPState, formData: FormData): Promise<OTPState> {
  const phone = toLatinDigits(String(formData.get("phone") ?? "").trim());
  const incoming = await headers();
  const response = await fetch(`${API_BASE}/api/v1/auth/otp/request`, {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      "X-Forwarded-For": clientIP(incoming),
    },
    body: JSON.stringify({ phone }),
    cache: "no-store",
  });
  if (!response.ok) {
    const err = await readError(response);
    return { error: err.message, phone };
  }
  const body = (await response.json()) as {
    phone_masked?: string;
    resend_after_seconds?: number;
  };
  return {
    sent: true,
    phone,
    phoneMasked: body.phone_masked,
    resendAfter: body.resend_after_seconds ?? 60,
  };
}

export async function verifyOTP(_prev: OTPState, formData: FormData): Promise<OTPState> {
  const phone = toLatinDigits(String(formData.get("phone") ?? "").trim());
  const code = toLatinDigits(String(formData.get("code") ?? "").trim());
  const incoming = await headers();
  const response = await fetch(`${API_BASE}/api/v1/auth/otp/verify`, {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      "X-Forwarded-For": clientIP(incoming),
    },
    body: JSON.stringify({ phone, code }),
    cache: "no-store",
  });
  if (!response.ok) {
    const err = await readError(response);
    return { sent: true, phone, error: err.message };
  }

  const jar = await cookies();
  let token = "";
  for (const line of response.headers.getSetCookie()) {
    const pair = line.split(";")[0];
    if (!pair) continue;
    const eq = pair.indexOf("=");
    if (eq < 0) continue;
    const name = pair.slice(0, eq).trim();
    const value = pair.slice(eq + 1);
    if (name === "user_session") {
      token = value;
      jar.set({
        name: "user_session",
        value,
        httpOnly: true,
        sameSite: "lax",
        path: "/",
        maxAge: 30 * 24 * 60 * 60,
        secure: process.env.NODE_ENV === "production",
      });
    }
  }

  const intent = String(formData.get("intent") ?? "");
  const productId = Number(formData.get("product_id") ?? 0);
  if (intent === "favorite" && productId > 0 && token) {
    await fetch(`${API_BASE}/api/v1/account/favorites`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        Cookie: `user_session=${token}`,
      },
      body: JSON.stringify({ product_id: productId }),
      cache: "no-store",
    });
  }

  const next = safeNextPath(String(formData.get("next") ?? "/account"));
  revalidatePath("/", "layout");
  redirect(next);
}

export async function addFavorite(productId: number): Promise<void> {
  const response = await accountMutate("/api/v1/account/favorites", {
    method: "POST",
    body: JSON.stringify({ product_id: productId }),
  });
  if (!response.ok) {
    throw await readError(response);
  }
  revalidatePath("/account");
  revalidatePath("/", "layout");
}

export async function removeFavorite(productId: number): Promise<void> {
  const response = await accountMutate(`/api/v1/account/favorites/${productId}`, {
    method: "DELETE",
  });
  if (!response.ok) {
    throw await readError(response);
  }
  revalidatePath("/account");
  revalidatePath("/", "layout");
}

export async function createAlert(formData: FormData): Promise<{ error?: string }> {
  const productId = Number(formData.get("product_id") ?? 0);
  const kind = String(formData.get("kind") ?? "price_drop");
  const rawTarget = String(formData.get("target_price_rial") ?? "").trim();
  const payload: { product_id: number; kind: string; target_price_rial?: number } = {
    product_id: productId,
    kind,
  };
  if (rawTarget !== "") {
    const n = Number(toLatinDigits(rawTarget));
    if (!Number.isFinite(n) || n < 0) {
      return { error: "قیمت هدف معتبر نیست." };
    }
    payload.target_price_rial = Math.round(n);
  }
  const response = await accountMutate("/api/v1/account/alerts", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  if (!response.ok) {
    return { error: (await readError(response)).message };
  }
  revalidatePath("/account");
  revalidatePath("/", "layout");
  return {};
}

export async function deleteAlert(id: number): Promise<void> {
  const response = await accountMutate(`/api/v1/account/alerts/${id}`, { method: "DELETE" });
  if (!response.ok) {
    throw await readError(response);
  }
  revalidatePath("/account");
  revalidatePath("/", "layout");
}

export async function clearSearches(): Promise<void> {
  const response = await accountMutate("/api/v1/account/searches", { method: "DELETE" });
  if (!response.ok) {
    throw await readError(response);
  }
  revalidatePath("/account");
  revalidatePath("/", "layout");
}
