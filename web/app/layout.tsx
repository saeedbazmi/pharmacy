import type { Metadata } from "next";
import localFont from "next/font/local";

import "./globals.css";

// Self-hosted so no request leaves for a third-party CDN.
const vazirmatn = localFont({
  src: [
    { path: "../public/fonts/Vazirmatn-Regular.woff2", weight: "400", style: "normal" },
    { path: "../public/fonts/Vazirmatn-Medium.woff2", weight: "500", style: "normal" },
    { path: "../public/fonts/Vazirmatn-Bold.woff2", weight: "700", style: "normal" },
  ],
  variable: "--font-vazirmatn",
  display: "swap",
  fallback: ["system-ui", "sans-serif"],
});

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000";

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  title: {
    default: "مقایسه قیمت دارو",
    template: "%s | مقایسه قیمت دارو",
  },
  description:
    "قیمت دارو و محصولات سلامت را در چند داروخانه مقایسه کنید و از داروخانه دلخواه خود خرید کنید.",
  openGraph: {
    type: "website",
    locale: "fa_IR",
    siteName: "مقایسه قیمت دارو",
  },
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="fa" dir="rtl" className={vazirmatn.variable}>
      <body className="min-h-screen antialiased">{children}</body>
    </html>
  );
}
