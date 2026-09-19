export default function CategoryLoading() {
  return (
    <main className="mx-auto flex max-w-5xl flex-col gap-8 px-6 py-10" aria-busy="true">
      <div className="h-8 w-48 rounded bg-border" />
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div className="h-56 rounded-xl bg-border" />
        <div className="h-56 rounded-xl bg-border" />
        <div className="h-56 rounded-xl bg-border" />
      </div>
    </main>
  );
}
