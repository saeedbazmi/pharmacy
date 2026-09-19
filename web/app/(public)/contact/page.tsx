import type { Metadata } from "next";

import { LegalPage, legalMetadata } from "@/components/layout/LegalPage";

export const revalidate = 60;

export function generateMetadata(): Promise<Metadata> {
  return legalMetadata("contact");
}

export default async function ContactPage() {
  return <LegalPage page="contact" />;
}
