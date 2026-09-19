import type { Metadata } from "next";

import { LegalPage, legalMetadata } from "@/components/layout/LegalPage";

export const revalidate = 60;

export function generateMetadata(): Promise<Metadata> {
  return legalMetadata("terms");
}

export default async function TermsPage() {
  return <LegalPage page="terms" />;
}
