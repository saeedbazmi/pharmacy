export const MAX_COMPARE = 4;
export const COMPARE_STORAGE_KEY = "pharmacy.compare.slugs";
export const COMPARE_CHANGE_EVENT = "pharmacy:compare-change";

const SLUG_PATTERN = /^[a-z0-9-]{1,200}$/;

export function parseCompareInput(raw: string | string[] | undefined): {
  slugs: string[];
  truncated: boolean;
} {
  const joined = Array.isArray(raw) ? raw.join(",") : (raw ?? "");
  const seen = new Set<string>();
  const valid: string[] = [];
  for (const part of joined.split(",")) {
    const slug = part.trim().toLowerCase();
    if (!slug || seen.has(slug) || !SLUG_PATTERN.test(slug)) continue;
    seen.add(slug);
    valid.push(slug);
  }
  return {
    slugs: valid.slice(0, MAX_COMPARE),
    truncated: valid.length > MAX_COMPARE,
  };
}

export function compareHref(slugs: string[]): string {
  if (slugs.length === 0) return "/compare";
  return `/compare?c=${encodeURIComponent(slugs.join(","))}`;
}

export function readCompareSlugs(): string[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = sessionStorage.getItem(COMPARE_STORAGE_KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parseCompareInput(parsed.filter((item) => typeof item === "string").join(",")).slugs;
  } catch {
    return [];
  }
}

export function writeCompareSlugs(slugs: string[]): void {
  const { slugs: clean } = parseCompareInput(slugs.join(","));
  sessionStorage.setItem(COMPARE_STORAGE_KEY, JSON.stringify(clean));
  window.dispatchEvent(new Event(COMPARE_CHANGE_EVENT));
}
