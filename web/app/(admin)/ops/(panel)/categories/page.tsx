import { CategoryManager } from "@/components/ops/CategoryManager";
import { opsRequest } from "@/lib/ops/server";
import type { OpsCategory } from "@/lib/ops/types";

export default async function CategoriesPage() {
  const body = await opsRequest<{ categories: OpsCategory[] }>("/api/v1/ops/categories");
  return (
    <main className="mx-auto flex max-w-3xl flex-col gap-6 px-6 py-8">
      <h1 className="text-2xl font-bold text-primary-dark">دسته‌بندی</h1>
      <CategoryManager categories={body.categories} />
    </main>
  );
}
