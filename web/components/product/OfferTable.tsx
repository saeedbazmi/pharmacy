import Link from "next/link";

import { Badge } from "@/components/ui/Badge";
import { ButtonLink } from "@/components/ui/Button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "@/components/ui/Table";
import { formatPrice, formatRelativeTime } from "@/lib/format";
import type { Offer } from "@/lib/api/types";

export function OfferTable({
  offers,
  filtered,
  resetHref,
}: {
  offers: Offer[];
  filtered?: boolean;
  resetHref?: string;
}) {
  if (offers.length === 0) {
    return (
      <div className="flex flex-col gap-3 rounded-xl border border-border bg-surface p-6 text-muted">
        <p>
          {filtered
            ? "هیچ پیشنهادی با این فیلتر پیدا نشد."
            : "این کالا در حال حاضر در هیچ داروخانه‌ای موجود نیست."}
        </p>
        {filtered && resetHref ? (
          <Link href={resetHref} className="text-primary hover:text-primary-dark">
            حذف فیلترها
          </Link>
        ) : (
          <Link href="/" className="text-primary hover:text-primary-dark">
            مشاهده کالاهای دیگر
          </Link>
        )}
      </div>
    );
  }

  return (
    <Table caption="قیمت این کالا در داروخانه‌ها">
      <TableHead>
        <TableRow>
          <TableHeaderCell>داروخانه</TableHeaderCell>
          <TableHeaderCell>قیمت</TableHeaderCell>
          <TableHeaderCell>موجودی</TableHeaderCell>
          <TableHeaderCell>به‌روزرسانی</TableHeaderCell>
          <TableHeaderCell>
            <span className="sr-only">خرید</span>
          </TableHeaderCell>
        </TableRow>
      </TableHead>
      <TableBody>
        {offers.map((offer) => (
          <TableRow key={offer.id}>
            <TableCell>{offer.pharmacy_name}</TableCell>
            <TableCell className="numeric">
              <div className="flex flex-wrap items-center gap-2">
                <span>{formatPrice(offer.price_rial)}</span>
                {offer.best_price ? (
                  <Badge tone="bestPrice">کم‌ترین قیمت</Badge>
                ) : null}
              </div>
            </TableCell>
            <TableCell>
              <StockStatus inStock={offer.in_stock} />
            </TableCell>
            <TableCell>
              <div className="flex flex-col gap-1 text-sm">
                <span className="text-muted">{formatRelativeTime(offer.last_seen_at)}</span>
                {offer.stale ? (
                  <Badge tone="warning">قیمت ممکن است قدیمی باشد — آخرین مشاهده بیش از یک روز پیش است</Badge>
                ) : null}
              </div>
            </TableCell>
            <TableCell>
              <ButtonLink
                href={`/go/${offer.id}`}
                variant="buy"
                size="sm"
                rel="nofollow sponsored"
                aria-label={`خرید ${offer.pharmacy_name} از سایت داروخانه`}
              >
                خرید از {offer.pharmacy_name}
              </ButtonLink>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

function StockStatus({ inStock }: { inStock: boolean }) {
  if (inStock) {
    return (
      <span className="inline-flex items-center gap-1.5 text-success">
        <InStockIcon />
        <span>موجود</span>
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1.5 text-danger">
      <OutOfStockIcon />
      <span>ناموجود</span>
    </span>
  );
}

function InStockIcon() {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 16 16"
      className="size-4 shrink-0"
      fill="none"
    >
      <circle cx="8" cy="8" r="7" stroke="currentColor" strokeWidth="1.5" />
      <path
        d="M4.5 8.2 7 10.7 11.5 5.5"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function OutOfStockIcon() {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 16 16"
      className="size-4 shrink-0"
      fill="none"
    >
      <circle cx="8" cy="8" r="7" stroke="currentColor" strokeWidth="1.5" />
      <path
        d="M5.5 5.5 10.5 10.5M10.5 5.5 5.5 10.5"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
    </svg>
  );
}
