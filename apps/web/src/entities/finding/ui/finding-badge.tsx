import type { Severity, FindingStatus } from "../model";

const SEVERITY_COLORS: Record<Severity, string> = {
  error: "bg-danger/15 text-danger",
  warning: "bg-warning/15 text-warning",
  note: "bg-primary/15 text-primary",
};

const STATUS_COLORS: Record<FindingStatus, string> = {
  new: "bg-danger/15 text-danger",
  existing: "bg-muted text-muted-foreground",
  resolved: "bg-success/15 text-success",
};

interface FindingBadgeProps {
  label: string;
  kind: "severity" | "status";
}

export function FindingBadge({ label, kind }: FindingBadgeProps) {
  const colorMap = kind === "severity" ? SEVERITY_COLORS : STATUS_COLORS;
  const colorClass = colorMap[label as keyof typeof colorMap] ?? "";

  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${colorClass}`}>
      {label}
    </span>
  );
}
