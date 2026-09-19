import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { ProductCard } from "@/components/product/ProductCard";
import { api, ApiError } from "@/lib/api/client";

export const revalidate = 300;

interface PageProps {
  params: Promise<{ slug: string }>;
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { slug } = await params;
  try {
    const { category } = await api.listByCategory(slug);
    return {
      title: category.name_fa,
      description: `قیمت کالاهای دسته ${category.name_fa} را در داروخانه‌های آنلاین مقایسه کنید.`,
      alternates: { canonical: `/category/${slug}` },
      openGraph: {
        title: category.name_fa,
        description: `مرور دسته ${category.name_fa}`,
        locale: "fa_IR",
        type: "website",
        url: `/category/${slug}`,
      },
    };
  } catch {
    return { title: "دسته یافت نشد" };
  }
}

export default async function CategoryPage({ params }: PageProps) {
  const { slug } = await params;
  let data;
  try {
    data = await api.listByCategory(slug);
  } catch (err) {
    if (err instanceof ApiError && err.isNotFound) {
      notFound();
    }
    throw err;
  }

  return (
    <main className="mx-auto flex max-w-5xl flex-col gap-8 px-6 py-10">
      <h1 className="text-2xl font-bold text-primary-dark">{data.category.name_fa}</h1>
      {data.products.length === 0 ? (
        <p className="rounded-xl border border-border bg-surface p-6 text-muted">
          در این دسته هنوز کالای منتشرشده‌ای نیست.
        </p>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {data.products.map((product) => (
            <ProductCard key={product.id} product={product} />
          ))}
        </div>
      )}
    </main>
  );
}
