import { SourceForm } from "@/components/ops/SourceForm";
import { SyncButton } from "@/components/ops/SyncButton";
import { opsRequest } from "@/lib/ops/server";
import type { OpsJob, OpsPharmacy, OpsSource } from "@/lib/ops/types";

export default async function SourceDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const [source, pharmacies, jobs] = await Promise.all([
    opsRequest<OpsSource>(`/api/v1/ops/sources/${id}`),
    opsRequest<{ pharmacies: OpsPharmacy[] }>("/api/v1/ops/pharmacies"),
    opsRequest<OpsJob>(`/api/v1/ops/sources/${id}/jobs`),
  ]);
  return (
    <main className="mx-auto flex max-w-3xl flex-col gap-6 px-6 py-8">
      <h1 className="text-2xl font-bold text-primary-dark">{source.pharmacy_name}</h1>
      <p className="text-sm text-muted">
        آخرین اجرا: {jobs.latest_run?.status ?? source.last_status ?? "هنوز اجرا نشده"}{" "}
        {jobs.latest_run?.error ? `— ${jobs.latest_run.error}` : ""}
      </p>
      {jobs.status ? <p className="text-sm">وضعیت job: {jobs.status}</p> : null}
      <SyncButton sourceId={source.id} />
      <SourceForm pharmacies={pharmacies.pharmacies} source={source} />
    </main>
  );
}
