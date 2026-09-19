import type { Metadata } from "next";

import { LegalPage, legalMetadata } from "@/components/layout/LegalPage";

export const revalidate = 60;

export function generateMetadata(): Promise<Metadata> {
  return legalMetadata("about");
}

export default async function AboutPage() {
  return <LegalPage page="about" />;
}
