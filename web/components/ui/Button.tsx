import type { ButtonHTMLAttributes, ReactNode } from "react";

/**
 * Base button. `buy` is the only variant allowed to use the accent colour and is
 * reserved for the "buy from pharmacy" call to action (AGENT.md section 8).
 */
type Variant = "primary" | "secondary" | "buy" | "ghost";
type Size = "sm" | "md";

const VARIANT_CLASSES: Record<Variant, string> = {
  primary: "bg-primary text-white hover:bg-primary-dark",
  secondary:
    "bg-surface text-primary-dark border border-border hover:bg-primary-soft",
  buy: "bg-accent text-white hover:bg-accent-dark",
  ghost: "bg-transparent text-primary-dark hover:bg-primary-soft",
};

const SIZE_CLASSES: Record<Size, string> = {
  sm: "px-3 py-1.5 text-sm",
  md: "px-4 py-2 text-base",
};

const BASE_CLASSES =
  "inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-60";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  children: ReactNode;
}

export function Button({
  variant = "primary",
  size = "md",
  className = "",
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      className={`${BASE_CLASSES} ${VARIANT_CLASSES[variant]} ${SIZE_CLASSES[size]} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}

interface ButtonLinkProps {
  href: string;
  variant?: Variant;
  size?: Size;
  className?: string;
  /** Outbound pharmacy links must not pass link equity. */
  rel?: string;
  target?: string;
  "aria-label"?: string;
  children: ReactNode;
}

/**
 * Anchor styled as a button. The buy action is a real link so it works without
 * JavaScript and can be opened in a new tab by the user.
 */
export function ButtonLink({
  href,
  variant = "primary",
  size = "md",
  className = "",
  rel,
  target,
  "aria-label": ariaLabel,
  children,
}: ButtonLinkProps) {
  return (
    <a
      href={href}
      rel={rel}
      target={target}
      aria-label={ariaLabel}
      className={`${BASE_CLASSES} ${VARIANT_CLASSES[variant]} ${SIZE_CLASSES[size]} ${className}`}
    >
      {children}
    </a>
  );
}
