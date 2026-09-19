import { SiteFooter } from "@/components/layout/SiteFooter";
import { SiteHeader } from "@/components/layout/SiteHeader";
import { api } from "@/lib/api/client";

export default async function PublicLayout({ children }: { children: React.ReactNode }) {
  const site = await api.siteSettings();
  return (
    <div className="flex min-h-screen flex-col">
      <SiteHeader />
      {children}
      <SiteFooter disclaimer={site.pages.disclaimer} />
    </div>
  );
}
