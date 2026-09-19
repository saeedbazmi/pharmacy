export default function ProductLoading() {
  return (
    <main className="mx-auto flex max-w-4xl flex-col gap-8 px-6 py-10" aria-busy="true">
      <div className="flex gap-6">
        <div className="size-48 shrink-0 rounded-xl bg-border" />
        <div className="flex flex-1 flex-col gap-3">
          <div className="h-8 w-2/3 rounded bg-border" />
          <div className="h-4 w-1/3 rounded bg-border" />
        </div>
      </div>
      <div className="overflow-hidden rounded-xl border border-border">
        <div className="grid grid-cols-5 gap-2 border-b border-border bg-surface p-3">
          <div className="h-4 rounded bg-border" />
          <div className="h-4 rounded bg-border" />
          <div className="h-4 rounded bg-border" />
          <div className="h-4 rounded bg-border" />
          <div className="h-4 rounded bg-border" />
        </div>
        <div className="grid grid-cols-5 gap-2 p-3">
          <div className="h-8 rounded bg-border" />
          <div className="h-8 rounded bg-border" />
          <div className="h-8 rounded bg-border" />
          <div className="h-8 rounded bg-border" />
          <div className="h-8 rounded bg-border" />
        </div>
        <div className="grid grid-cols-5 gap-2 p-3">
          <div className="h-8 rounded bg-border" />
          <div className="h-8 rounded bg-border" />
          <div className="h-8 rounded bg-border" />
          <div className="h-8 rounded bg-border" />
          <div className="h-8 rounded bg-border" />
        </div>
      </div>
    </main>
  );
}
