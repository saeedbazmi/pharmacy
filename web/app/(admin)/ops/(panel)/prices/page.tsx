import { PriceReviewQueue } from "@/components/ops/PriceReviewQueue";
import { opsRequest } from "@/lib/ops/server";
import type { OpsSuspiciousOffer } from "@/lib/ops/types";

export default async function PricesPage({
  searchParams,
}: {
  searchParams: Promise<{ page?: string }>;
}) {
  const q = await searchParams;
  const page = q.page ?? "1";
  const body = await opsRequest<{
    offers: OpsSuspiciousOffer[];
    page: number;
    page_size: number;
    total: number;
  }>(`/api/v1/ops/prices?page=${page}`);
  return (
    <main className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-8">
      <h1 className="text-2xl font-bold text-primary-dark">صف بازبینی قیمت</h1>
      <p className="text-muted">
        جهش بیش از آستانه منتشر نمی‌شود. قیمت قبلی روی سایت می‌ماند تا قیمت جدید را
        تایید یا رد کنید.
      </p>
      <PriceReviewQueue offers={body.offers} total={body.total} page={body.page} />
    </main>
  );
}
