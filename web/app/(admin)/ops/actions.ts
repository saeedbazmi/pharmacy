"use server";

import { revalidateTag } from "next/cache";

export async function revalidateCatalog(slug?: string) {
  revalidateTag("products");
  revalidateTag("search");
  revalidateTag("sitemap");
  if (slug) {
    revalidateTag(`product:${slug}`);
  }
}

export async function revalidateSite() {
  revalidateTag("site");
  revalidateTag("sitemap");
}
