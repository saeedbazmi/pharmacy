import { ProductEditForm } from "@/components/ops/ProductEditForm";
import { opsRequest } from "@/lib/ops/server";
import type { OpsBrand, OpsCategory, OpsProduct } from "@/lib/ops/types";

export default async function OpsProductPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const [product, brands, categories] = await Promise.all([
    opsRequest<OpsProduct>(`/api/v1/ops/products/${id}`),
    opsRequest<{ brands: OpsBrand[] }>("/api/v1/ops/brands"),
    opsRequest<{ categories: OpsCategory[] }>("/api/v1/ops/categories"),
  ]);
  return (
    <main className="mx-auto flex max-w-3xl flex-col gap-6 px-6 py-8">
      <h1 className="text-2xl font-bold text-primary-dark">{product.name_fa}</h1>
      <ProductEditForm product={product} brands={brands.brands} categories={categories.categories} />
    </main>
  );
}
