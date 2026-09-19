"use client";

import { useRouter } from "next/navigation";
import type { FormEvent } from "react";

import { Button } from "@/components/ui/Button";
import { normalizeQuery } from "@/lib/persian";

export function SearchBox({
  defaultQuery = "",
  className = "",
}: {
  defaultQuery?: string;
  className?: string;
}) {
  const router = useRouter();

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    const form = event.currentTarget;
    const data = new FormData(form);
    const q = normalizeQuery(String(data.get("q") ?? ""));
    if (!q) {
      event.preventDefault();
      return;
    }
    // Prefer a normalised URL when JS is available; without JS the GET form still works.
    event.preventDefault();
    const params = new URLSearchParams();
    params.set("q", q);
    router.push(`/search?${params.toString()}`);
  }

  return (
    <form
      action="/search"
      method="get"
      onSubmit={onSubmit}
      className={`flex flex-wrap items-end gap-2 ${className}`}
      role="search"
    >
      <label htmlFor="search-q" className="flex min-w-48 flex-1 flex-col gap-1 text-sm">
        <span className="text-muted">جست‌وجوی دارو</span>
        <input
          id="search-q"
          name="q"
          type="search"
          defaultValue={defaultQuery}
          autoComplete="off"
          enterKeyHint="search"
          placeholder="نام دارو، برند یا ماده مؤثره"
          className="rounded-lg border border-border bg-background px-3 py-2"
        />
      </label>
      <Button type="submit" variant="primary" size="md">
        جست‌وجو
      </Button>
    </form>
  );
}
