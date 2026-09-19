import { BrandManager } from "@/components/ops/BrandManager";
import { opsRequest } from "@/lib/ops/server";
import type { OpsBrand } from "@/lib/ops/types";

export default async function BrandsPage() {
  const body = await opsRequest<{ brands: OpsBrand[] }>("/api/v1/ops/brands");
  return (
    <main className="mx-auto flex max-w-3xl flex-col gap-6 px-6 py-8">
      <h1 className="text-2xl font-bold text-primary-dark">برندها</h1>
      <BrandManager brands={body.brands} />
    </main>
  );
}
