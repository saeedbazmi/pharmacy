"use client";

import { addFavorite, removeFavorite } from "@/app/(public)/account/actions";
import { SubmitButton } from "@/components/account/SubmitButton";

export function FavoriteButton({
  productId,
  slug,
  signedIn,
  isFavorite,
}: {
  productId: number;
  slug: string;
  signedIn: boolean;
  isFavorite: boolean;
}) {
  if (!signedIn) {
    const params = new URLSearchParams({
      next: `/product/${slug}`,
      intent: "favorite",
      product_id: String(productId),
    });
    return (
      <a
        href={`/login?${params.toString()}`}
        className="inline-flex items-center justify-center rounded-lg border border-border bg-surface px-4 py-2 text-sm font-medium text-primary-dark hover:bg-primary-soft"
      >
        افزودن به علاقه‌مندی
      </a>
    );
  }

  if (isFavorite) {
    return (
      <form action={removeFavorite.bind(null, productId)}>
        <SubmitButton pendingLabel="در حال حذف…">حذف از علاقه‌مندی</SubmitButton>
      </form>
    );
  }

  return (
    <form action={addFavorite.bind(null, productId)}>
      <SubmitButton pendingLabel="در حال ذخیره…">افزودن به علاقه‌مندی</SubmitButton>
    </form>
  );
}
