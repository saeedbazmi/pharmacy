"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";

import { Button } from "@/components/ui/Button";
import { formatPrice } from "@/lib/format";
import type { OpsSuspiciousOffer } from "@/lib/ops/types";
import { revalidateCatalog } from "@/app/(admin)/ops/actions";

export function PriceReviewQueue({
  offers,
  total,
  page,
}: {
  offers: OpsSuspiciousOffer[];
  total: number;
  page: number;
}) {
  const router = useRouter();
  const [error, setError] = useState("");

  async function act(offer: OpsSuspiciousOffer, action: "approve" | "reject") {
    const label = action === "approve" ? "تایید قیمت جدید؟" : "رد شود و قیمت قبلی بماند؟";
    if (!confirm(label)) return;
    const response = await fetch(`/api/v1/ops/prices/${offer.id}/${action}`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: "{}",
    });
    if (!response.ok) {
      const err = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
      setError(err?.error?.message ?? "عملیات ناموفق بود.");
      return;
    }
    setError("");
    await revalidateCatalog(offer.product_slug);
    router.refresh();
  }

  return (
    <div className="flex flex-col gap-4">
      <p className="text-sm text-muted">{total} مورد در صف</p>
      {error ? <p className="text-sm text-danger">{error}</p> : null}
      {offers.length === 0 ? <p className="text-muted">صف خالی است.</p> : null}
      <ul className="flex flex-col gap-4">
        {offers.map((offer) => (
          <li key={offer.id} className="rounded-xl border border-border bg-surface p-4">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <Link href={`/product/${offer.product_slug}`} className="font-medium text-primary hover:text-primary-dark">
                  {offer.product_name}
                </Link>
                <div className="mt-1 text-sm text-muted">{offer.pharmacy_name}</div>
              </div>
              <a
                href={offer.product_url}
                target="_blank"
                rel="noreferrer"
                className="text-sm text-primary hover:text-primary-dark"
              >
                صفحه منبع
              </a>
            </div>
            <dl className="mt-3 grid gap-2 text-sm sm:grid-cols-2">
              <div>
                <dt className="text-muted">قیمت فعلی (منتشرشده)</dt>
                <dd className="font-medium">{formatPrice(offer.current_price_rial)}</dd>
              </div>
              <div>
                <dt className="text-muted">قیمت پیشنهادی منبع</dt>
                <dd className="font-medium text-warning">{formatPrice(offer.proposed_price_rial)}</dd>
              </div>
            </dl>
            {offer.source_reject_count > 0 ? (
              <p className="mt-2 text-xs text-warning">ردهای قبلی این منبع: {offer.source_reject_count}</p>
            ) : null}
            <div className="mt-3 flex flex-wrap gap-2">
              <Button type="button" size="sm" onClick={() => act(offer, "approve")}>
                تایید قیمت جدید
              </Button>
              <Button type="button" size="sm" variant="ghost" onClick={() => act(offer, "reject")}>
                رد و نگه‌داشتن قیمت قبلی
              </Button>
            </div>
          </li>
        ))}
      </ul>
      {total > 20 ? (
        <p className="text-sm">
          صفحه {page}
          {page > 1 ? (
            <Link href={`/ops/prices?page=${page - 1}`} className="mr-3 text-primary">
              قبلی
            </Link>
          ) : null}
          {page * 20 < total ? (
            <Link href={`/ops/prices?page=${page + 1}`} className="mr-3 text-primary">
              بعدی
            </Link>
          ) : null}
        </p>
      ) : null}
    </div>
  );
}
