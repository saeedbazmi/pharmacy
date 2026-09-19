"use client";

import { useActionState } from "react";

import { createAlert } from "@/app/(public)/account/actions";
import { SubmitButton } from "@/components/account/SubmitButton";

const initial = {} as { error?: string };

export function AlertForm({
  productId,
  slug,
  signedIn,
}: {
  productId: number;
  slug: string;
  signedIn: boolean;
}) {
  const [state, action] = useActionState(
    async (_prev: { error?: string }, formData: FormData) => createAlert(formData),
    initial,
  );

  if (!signedIn) {
    const params = new URLSearchParams({
      next: `/product/${slug}`,
      intent: "alert",
    });
    return (
      <p className="text-sm text-muted">
        برای ثبت هشدار قیمت{" "}
        <a href={`/login?${params.toString()}`} className="text-primary hover:text-primary-dark">
          وارد شوید
        </a>
        . جست‌وجو و خرید بدون ورود ممکن است.
      </p>
    );
  }

  return (
    <form action={action} className="flex flex-col gap-3 rounded-lg border border-border bg-surface p-4">
      <h3 className="font-medium">هشدار قیمت یا موجودی</h3>
      <input type="hidden" name="product_id" value={productId} />
      <label className="flex flex-col gap-1 text-sm" htmlFor="alert-kind">
        نوع هشدار
        <select
          id="alert-kind"
          name="kind"
          className="rounded-lg border border-border px-3 py-2"
          defaultValue="price_drop"
        >
          <option value="price_drop">کاهش قیمت</option>
          <option value="back_in_stock">موجود شدن</option>
        </select>
      </label>
      <label className="flex flex-col gap-1 text-sm" htmlFor="target-price">
        قیمت هدف (اختیاری، ریال)
        <input
          id="target-price"
          name="target_price_rial"
          type="text"
          inputMode="numeric"
          className="rounded-lg border border-border px-3 py-2"
        />
      </label>
      {state.error ? (
        <p className="text-sm text-danger" role="alert">
          {state.error}
        </p>
      ) : null}
      <SubmitButton pendingLabel="در حال ثبت…">ثبت هشدار</SubmitButton>
    </form>
  );
}
