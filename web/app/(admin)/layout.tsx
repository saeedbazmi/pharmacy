import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "پنل داده",
  robots: { index: false, follow: false },
};

export default function AdminRootLayout({ children }: { children: React.ReactNode }) {
  return children;
}
