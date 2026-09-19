"use client";

import { useEffect, useState } from "react";

import {
  COMPARE_CHANGE_EVENT,
  compareHref,
  readCompareSlugs,
} from "@/lib/compare";
import { formatNumber } from "@/lib/format";

export function CompareNav() {
  const [count, setCount] = useState(0);
  const [href, setHref] = useState("/compare");

  useEffect(() => {
    const sync = () => {
      const slugs = readCompareSlugs();
      setCount(slugs.length);
      setHref(compareHref(slugs));
    };
    sync();
    window.addEventListener(COMPARE_CHANGE_EVENT, sync);
    return () => window.removeEventListener(COMPARE_CHANGE_EVENT, sync);
  }, []);

  if (count === 0) {
    return (
      <a href="/compare" className="text-sm text-primary-dark hover:text-primary">
        مقایسه کالاها
      </a>
    );
  }

  return (
    <a href={href} className="text-sm text-primary-dark hover:text-primary">
      مشاهده مقایسه ({formatNumber(count)})
    </a>
  );
}
