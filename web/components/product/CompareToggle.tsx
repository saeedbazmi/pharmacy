"use client";

import { useEffect, useId, useState } from "react";

import {
  COMPARE_CHANGE_EVENT,
  MAX_COMPARE,
  readCompareSlugs,
  writeCompareSlugs,
} from "@/lib/compare";

export function CompareToggle({ slug, name }: { slug: string; name: string }) {
  const id = useId();
  const [selected, setSelected] = useState(false);
  const [limitReached, setLimitReached] = useState(false);

  useEffect(() => {
    const sync = () => {
      const slugs = readCompareSlugs();
      setSelected(slugs.includes(slug));
      setLimitReached(slugs.length >= MAX_COMPARE && !slugs.includes(slug));
    };
    sync();
    window.addEventListener(COMPARE_CHANGE_EVENT, sync);
    window.addEventListener("storage", sync);
    return () => {
      window.removeEventListener(COMPARE_CHANGE_EVENT, sync);
      window.removeEventListener("storage", sync);
    };
  }, [slug]);

  function onChange(checked: boolean) {
    const slugs = readCompareSlugs();
    if (checked) {
      if (slugs.includes(slug)) return;
      if (slugs.length >= MAX_COMPARE) {
        setLimitReached(true);
        return;
      }
      writeCompareSlugs([...slugs, slug]);
      return;
    }
    writeCompareSlugs(slugs.filter((item) => item !== slug));
    setLimitReached(false);
  }

  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={id} className="inline-flex cursor-pointer items-center gap-2 text-sm">
        <input
          id={id}
          type="checkbox"
          checked={selected}
          disabled={limitReached}
          onChange={(event) => onChange(event.target.checked)}
          className="size-4 accent-primary"
        />
        <span>{selected ? "در مقایسه" : "افزودن به مقایسه"}</span>
      </label>
      {limitReached ? (
        <p className="text-xs text-warning" role="status">
          حداکثر {MAX_COMPARE} کالا می‌توانید مقایسه کنید. یکی را از مقایسه حذف
          کنید تا «{name}» اضافه شود.
        </p>
      ) : null}
    </div>
  );
}
