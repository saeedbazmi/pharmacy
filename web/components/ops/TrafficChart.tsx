"use client";

export function TrafficChart({
  series,
}: {
  series: { day: string; searches: number; clicks: number }[];
}) {
  if (series.length === 0) {
    return <p className="text-sm text-muted">در این بازه نموداری نیست.</p>;
  }
  const max = Math.max(1, ...series.map((p) => Math.max(p.searches, p.clicks)));
  const width = 640;
  const height = 192;
  const gap = 8;
  const barW = Math.max(4, (width - gap * (series.length + 1)) / series.length / 2);
  return (
    <svg
      role="img"
      aria-label="نمودار جست‌وجو و کلیک روزانه"
      viewBox={`0 0 ${width} ${height}`}
      className="h-48 w-full"
    >
      {series.map((p, i) => {
        const x = gap + i * ((width - gap) / series.length);
        const searchH = (p.searches / max) * (height - 24);
        const clickH = (p.clicks / max) * (height - 24);
        return (
          <g key={p.day}>
            <rect
              x={x}
              y={height - 16 - searchH}
              width={barW}
              height={searchH}
              className="fill-primary"
            />
            <rect
              x={x + barW + 2}
              y={height - 16 - clickH}
              width={barW}
              height={clickH}
              className="fill-primary-dark"
            />
          </g>
        );
      })}
    </svg>
  );
}
