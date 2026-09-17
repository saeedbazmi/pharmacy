import type { Metadata } from "next";

import { Badge } from "@/components/ui/Badge";
import { Button, ButtonLink } from "@/components/ui/Button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "@/components/ui/Table";
import { formatPrice, formatRelativeTime } from "@/lib/format";

/** Internal review page for the design tokens. Never indexed, never linked publicly. */
export const metadata: Metadata = {
  title: "توکن‌های ظاهری",
  robots: { index: false, follow: false },
};

const COLORS = [
  { name: "primary", usage: "برند، هدر، لینک فعال", className: "bg-primary" },
  { name: "primary-dark", usage: "hover و متن روی سطح روشن", className: "bg-primary-dark" },
  { name: "primary-soft", usage: "پس‌زمینه بج و حالت انتخاب", className: "bg-primary-soft" },
  { name: "accent", usage: "فقط دکمه خرید از داروخانه", className: "bg-accent" },
  { name: "best-price", usage: "نشان کم‌ترین قیمت", className: "bg-best-price" },
  { name: "success", usage: "موجود", className: "bg-success" },
  { name: "warning", usage: "داده کهنه", className: "bg-warning" },
  { name: "danger", usage: "ناموجود و خطا", className: "bg-danger" },
  { name: "info", usage: "راهنما و اطلاعیه", className: "bg-info" },
  { name: "border", usage: "خط جداکننده", className: "bg-border" },
  { name: "background", usage: "پس‌زمینه صفحه", className: "bg-background" },
] as const;

const SAMPLE_UPDATED_AT = new Date(Date.now() - 12 * 60 * 1000).toISOString();

export default function TokensPage() {
  return (
    <main className="mx-auto flex max-w-4xl flex-col gap-10 px-6 py-12">
      <header>
        <h1 className="text-2xl font-bold text-primary-dark">توکن‌های ظاهری</h1>
        <p className="mt-2 text-sm text-muted">
          صفحه بازبینی داخلی. هر رنگی که در کامپوننت‌ها استفاده می‌شود باید از این
          فهرست بیاید.
        </p>
      </header>

      <section className="flex flex-col gap-4">
        <h2 className="text-lg font-medium">پالت رنگ</h2>
        <ul className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          {COLORS.map((color) => (
            <li
              key={color.name}
              className="flex items-center gap-3 rounded-lg border border-border bg-surface p-3"
            >
              <span
                className={`size-10 shrink-0 rounded-md border border-border ${color.className}`}
                aria-hidden="true"
              />
              <span className="flex flex-col">
                <code className="text-sm">{color.name}</code>
                <span className="text-xs text-muted">{color.usage}</span>
              </span>
            </li>
          ))}
        </ul>
      </section>

      <section className="flex flex-col gap-4">
        <h2 className="text-lg font-medium">دکمه‌ها</h2>
        <div className="flex flex-wrap items-center gap-3">
          <Button>دکمه اصلی</Button>
          <Button variant="secondary">دکمه ثانویه</Button>
          <Button variant="ghost">دکمه ساده</Button>
          <ButtonLink href="#" variant="buy" rel="nofollow sponsored">
            خرید از داروخانه
          </ButtonLink>
          <Button disabled>غیرفعال</Button>
        </div>
      </section>

      <section className="flex flex-col gap-4">
        <h2 className="text-lg font-medium">بج‌ها</h2>
        <div className="flex flex-wrap items-center gap-3">
          <Badge tone="bestPrice">کم‌ترین قیمت</Badge>
          <Badge tone="success">موجود</Badge>
          <Badge tone="danger">ناموجود</Badge>
          <Badge tone="warning">قیمت کهنه</Badge>
          <Badge tone="info">اطلاعیه</Badge>
          <Badge>عادی</Badge>
        </div>
      </section>

      <section className="flex flex-col gap-4">
        <h2 className="text-lg font-medium">جدول و قالب‌بندی</h2>
        <Table caption="نمونه جدول قیمت داروخانه‌ها">
          <TableHead>
            <TableRow>
              <TableHeaderCell>داروخانه</TableHeaderCell>
              <TableHeaderCell>قیمت</TableHeaderCell>
              <TableHeaderCell>موجودی</TableHeaderCell>
              <TableHeaderCell>به‌روزرسانی</TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            <TableRow>
              <TableCell>داروخانه نمونه یک</TableCell>
              <TableCell className="numeric">
                {formatPrice(125_000)} <Badge tone="bestPrice">کم‌ترین قیمت</Badge>
              </TableCell>
              <TableCell>
                <Badge tone="success">موجود</Badge>
              </TableCell>
              <TableCell className="text-sm text-muted">
                {formatRelativeTime(SAMPLE_UPDATED_AT)}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>داروخانه نمونه دو</TableCell>
              <TableCell className="numeric">{formatPrice(148_500)}</TableCell>
              <TableCell>
                <Badge tone="danger">ناموجود</Badge>
              </TableCell>
              <TableCell className="text-sm text-muted">
                {formatRelativeTime(SAMPLE_UPDATED_AT)}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </section>
    </main>
  );
}
