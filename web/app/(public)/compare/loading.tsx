export default function CompareLoading() {
  return (
    <main className="mx-auto flex max-w-4xl flex-col gap-6 px-6 py-10" aria-busy="true">
      <div className="h-8 w-48 rounded bg-border" />
      <div className="h-40 rounded-xl bg-border" />
    </main>
  );
}
