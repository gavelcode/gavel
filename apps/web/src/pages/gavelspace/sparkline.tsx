interface SparklinePoint {
  day: string;
  count: number;
}

interface SparklineProps {
  series: SparklinePoint[];
  width?: number;
  height?: number;
}

export function Sparkline({ series, width = 200, height = 40 }: SparklineProps) {
  if (series.length === 0) {
    return (
      <svg
        role="img"
        aria-label="No activity yet"
        width={width}
        height={height}
        viewBox={`0 0 ${width} ${height}`}
        className="text-muted-foreground"
      />
    );
  }

  const counts = series.map((p) => p.count);
  const max = Math.max(...counts, 1);
  const stepX = series.length > 1 ? width / (series.length - 1) : 0;

  const points = counts
    .map((c, i) => {
      const x = i * stepX;
      const y = height - (c / max) * (height - 4) - 2;
      return `${x.toFixed(1)},${y.toFixed(1)}`;
    })
    .join(" ");

  return (
    <svg
      role="img"
      aria-label="7-day findings sparkline"
      width={width}
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      className="text-muted-foreground"
    >
      <polyline
        points={points}
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinejoin="round"
      />
    </svg>
  );
}
