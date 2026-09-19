import { SourceForm } from "@/components/ops/SourceForm";
import { opsRequest } from "@/lib/ops/server";
import type { OpsPharmacy } from "@/lib/ops/types";

export default async function NewSourcePage() {
  const body = await opsRequest<{ pharmacies: OpsPharmacy[] }>("/api/v1/ops/pharmacies");
  return (
    <main className="mx-auto flex max-w-3xl flex-col gap-6 px-6 py-8">
      <h1 className="text-2xl font-bold text-primary-dark">اتصال داروخانه تازه</h1>
      <p className="text-sm text-muted">
        نوع منبع را از فهرست موجود انتخاب کنید (داروکده یا روشا). نشانی باید
        عمومی و http/https باشد. کلید API فقط نوشته می‌شود و بعداً ماسک می‌ماند.
      </p>
      <SourceForm pharmacies={body.pharmacies} />
    </main>
  );
}
