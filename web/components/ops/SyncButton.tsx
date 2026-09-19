"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/Button";

export function SyncButton({ sourceId }: { sourceId: number }) {
  const router = useRouter();
  const [message, setMessage] = useState("");
  const [pending, setPending] = useState(false);

  async function run() {
    setPending(true);
    const response = await fetch(`/api/v1/ops/sources/${sourceId}/sync`, {
      method: "POST",
      credentials: "include",
      headers: { Accept: "application/json" },
    });
    setPending(false);
    if (!response.ok) {
      setMessage("ثبت سینک ناموفق بود.");
      return;
    }
    const body = (await response.json()) as { enqueued: boolean; status?: string };
    setMessage(body.enqueued ? "job ثبت شد." : "همین job باز است؛ تکراری ساخته نشد.");
    router.refresh();
  }

  return (
    <div className="flex items-center gap-3">
      <Button type="button" variant="secondary" onClick={run} disabled={pending}>
        سینک همین حالا
      </Button>
      {message ? <span className="text-sm text-muted">{message}</span> : null}
    </div>
  );
}
