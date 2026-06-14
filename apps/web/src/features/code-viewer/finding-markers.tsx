import { AlertOctagon, AlertTriangle, Info } from "lucide-react";
import type { Finding, Severity } from "@/entities/finding/model";

export interface FindingMarkersProps {
  findings: Finding[];
  onFindingClick?: (finding: Finding) => void;
}

const severityClass: Record<Severity, string> = {
  error: "text-danger hover:bg-danger/10",
  warning: "text-warning hover:bg-warning/10",
  note: "text-muted-foreground hover:bg-muted/40",
};

const severityLabel: Record<Severity, string> = {
  error: "error",
  warning: "warning",
  note: "note",
};

const VISIBLE_LIMIT = 3;

function IconFor({ severity }: { severity: Severity }) {
  const className = "h-3.5 w-3.5";
  if (severity === "error") return <AlertOctagon className={className} aria-hidden="true" />;
  if (severity === "warning") return <AlertTriangle className={className} aria-hidden="true" />;
  return <Info className={className} aria-hidden="true" />;
}

export function FindingMarkers({ findings, onFindingClick }: FindingMarkersProps) {
  if (findings.length === 0) {
    return <span className="block w-6" aria-hidden="true" />;
  }

  const visible = findings.slice(0, VISIBLE_LIMIT);
  const overflow = findings.length - visible.length;

  return (
    <span className="flex items-center gap-0.5 pl-1 pr-2">
      {visible.map((finding) => (
        <button
          key={finding.fingerprint}
          type="button"
          data-testid="finding-marker"
          data-severity={finding.severity}
          aria-label={`${severityLabel[finding.severity]}: ${finding.ruleId} — ${finding.message}`}
          onClick={() => onFindingClick?.(finding)}
          className={`inline-flex h-5 w-5 items-center justify-center rounded ${severityClass[finding.severity]}`}
        >
          <IconFor severity={finding.severity} />
        </button>
      ))}
      {overflow > 0 && (
        <span
          data-testid="finding-marker-overflow"
          className="inline-flex h-5 items-center rounded bg-muted/60 px-1 text-[10px] font-semibold text-muted-foreground tabular-nums"
          aria-label={`${overflow} more findings on this line`}
        >
          +{overflow}
        </span>
      )}
    </span>
  );
}
