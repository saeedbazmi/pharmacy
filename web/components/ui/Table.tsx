import type { ReactNode, ThHTMLAttributes } from "react";

/**
 * Thin wrappers over real table elements. The price comparison table must be a
 * genuine <table> with scoped headers so screen readers can read it.
 */

export function Table({
  caption,
  children,
}: {
  /** Visually hidden description of what the table contains. */
  caption?: string;
  children: ReactNode;
}) {
  return (
    <div className="overflow-x-auto rounded-xl border border-border bg-surface">
      <table className="w-full border-collapse text-start">
        {caption ? <caption className="sr-only">{caption}</caption> : null}
        {children}
      </table>
    </div>
  );
}

export function TableHead({ children }: { children: ReactNode }) {
  return <thead className="bg-background text-muted">{children}</thead>;
}

export function TableBody({ children }: { children: ReactNode }) {
  return <tbody>{children}</tbody>;
}

export function TableRow({ children }: { children: ReactNode }) {
  return <tr className="border-t border-border">{children}</tr>;
}

interface HeaderCellProps extends ThHTMLAttributes<HTMLTableCellElement> {
  children: ReactNode;
}

export function TableHeaderCell({
  scope = "col",
  className = "",
  children,
  ...props
}: HeaderCellProps) {
  return (
    <th
      scope={scope}
      className={`px-4 py-3 text-start text-sm font-medium ${className}`}
      {...props}
    >
      {children}
    </th>
  );
}

export function TableCell({
  className = "",
  children,
}: {
  className?: string;
  children: ReactNode;
}) {
  return <td className={`px-4 py-3 text-start ${className}`}>{children}</td>;
}
