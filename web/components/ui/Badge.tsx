import type { ReactNode } from "react";

/**
 * Small status label. Colour never carries the meaning on its own, so every
 * badge renders text; callers must not use a badge as a bare colour dot.
 */
type Tone = "neutral" | "bestPrice" | "success" | "warning" | "danger" | "info";

const TONE_CLASSES: Record<Tone, string> = {
  neutral: "bg-background text-muted border-border",
  bestPrice: "bg-primary-soft text-best-price border-best-price",
  success: "bg-primary-soft text-success border-success",
  warning: "bg-background text-warning border-warning",
  danger: "bg-background text-danger border-danger",
  info: "bg-background text-info border-info",
};

interface BadgeProps {
  tone?: Tone;
  className?: string;
  children: ReactNode;
}

export function Badge({ tone = "neutral", className = "", children }: BadgeProps) {
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs font-medium ${TONE_CLASSES[tone]} ${className}`}
    >
      {children}
    </span>
  );
}
