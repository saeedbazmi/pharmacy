import { MatchQueue } from "@/components/ops/MatchQueue";
import { opsRequest } from "@/lib/ops/server";
import type { OpsMatch, OpsSource } from "@/lib/ops/types";

export default async function MatchesPage({
  searchParams,
}: {
  searchParams: Promise<{ source_id?: string; min_score?: string; page?: string }>;
}) {
  const q = await searchParams;
  const params = new URLSearchParams();
  if (q.source_id) params.set("source_id", q.source_id);
  if (q.min_score) params.set("min_score", q.min_score);
  if (q.page) params.set("page", q.page);
  params.set("page_size", "20");
  const [matches, sources] = await Promise.all([
    opsRequest<{ matches: OpsMatch[]; total: number; page: number; page_size: number }>(
      `/api/v1/ops/matches?${params.toString()}`,
    ),
    opsRequest<{ sources: OpsSource[] }>("/api/v1/ops/sources"),
  ]);
  return (
    <main className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-8">
      <h1 className="text-2xl font-bold text-primary-dark">صف تطبیق</h1>
      <MatchQueue
        matches={matches.matches}
        total={matches.total}
        page={matches.page}
        sources={sources.sources}
        sourceId={q.source_id ?? ""}
        minScore={q.min_score ?? ""}
      />
    </main>
  );
}
