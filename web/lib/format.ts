/**
 * Display formatting. Values are stored and transmitted as Latin integers in
 * rial; Persian digits exist only on screen.
 */

const numberFormatter = new Intl.NumberFormat("fa-IR");
const relativeFormatter = new Intl.RelativeTimeFormat("fa", { numeric: "auto" });
const dateFormatter = new Intl.DateTimeFormat("fa-IR", {
  dateStyle: "medium",
});

/** Formats an integer rial amount, e.g. 125000 -> "۱۲۵٬۰۰۰ ریال". */
export function formatPrice(rial: number): string {
  return `${numberFormatter.format(Math.round(rial))} ریال`;
}

/** Formats any number with Persian digits and thousand separators. */
export function formatNumber(value: number): string {
  return numberFormatter.format(value);
}

/** Formats a date in the Persian calendar, e.g. "۲۶ شهریور ۱۴۰۵". */
export function formatDate(value: string | Date): string {
  return dateFormatter.format(toDate(value));
}

const MINUTE = 60;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

/**
 * Formats price freshness the way the product page shows it, e.g. "۱۲ دقیقه
 * پیش". Freshness is the trust signal of this product, so it is always visible.
 */
export function formatRelativeTime(
  value: string | Date,
  now: Date = new Date(),
): string {
  const seconds = Math.round((toDate(value).getTime() - now.getTime()) / 1000);
  const magnitude = Math.abs(seconds);

  if (magnitude < MINUTE) return "همین حالا";
  if (magnitude < HOUR) {
    return relativeFormatter.format(Math.round(seconds / MINUTE), "minute");
  }
  if (magnitude < DAY) {
    return relativeFormatter.format(Math.round(seconds / HOUR), "hour");
  }
  if (magnitude < 30 * DAY) {
    return relativeFormatter.format(Math.round(seconds / DAY), "day");
  }
  return formatDate(value);
}

function toDate(value: string | Date): Date {
  return value instanceof Date ? value : new Date(value);
}
