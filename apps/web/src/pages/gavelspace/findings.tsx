import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import type { Finding, FindingFilters, Severity as SeverityType } from "@/entities/finding/model";
import * as findingApi from "@/entities/finding/api";
import { Badge } from "@/shared/ui/badge";
import { Severity } from "@/shared/ui/severity";
import { Card } from "@/shared/ui/card";
import { Spinner } from "@/shared/ui/spinner";
import { EmptyState } from "@/shared/ui/empty-state";
import { cn } from "@/shared/lib/utils";

const SEVERITY_OPTIONS: SeverityType[] = ["error", "warning", "note"];

function IssueRow({
  finding,
  active,
  gavelspaceName,
  onClick,
}: {
  finding: Finding;
  active: boolean;
  gavelspaceName: string;
  onClick: () => void;
}) {
  return (
    <Link
      to={`/gavelspaces/${gavelspaceName}/findings/${finding.fingerprint}`}
      onClick={onClick}
      className={cn(
        "block w-full border-b border-border px-3.5 py-3 text-left transition-colors duration-fast focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-inset",
        active
          ? "border-l-2 border-l-primary bg-primary/5"
          : "border-l-2 border-l-transparent hover:bg-muted/50",
      )}
    >
      <div className="mb-1.5 flex items-center gap-2">
        <Severity level={finding.severity} />
        <span className="font-mono text-2xs text-muted-foreground">{finding.tool}</span>
      </div>
      <div className={cn("mb-1 text-xs", active ? "font-medium" : "")}>{finding.message}</div>
      <div className="truncate font-mono text-xs text-muted-foreground">
        {finding.filePath}:{finding.line}
      </div>
    </Link>
  );
}

function IssueDetail({ finding }: { finding: Finding }) {
  return (
    <div className="flex flex-col gap-4 overflow-auto p-6">
      <div>
        <div className="mb-2 flex flex-wrap items-center gap-2">
          <Severity level={finding.severity} />
          <Badge tone="neutral">{finding.ruleId}</Badge>
          <Badge tone="neutral">{finding.tool}</Badge>
          <Badge tone={finding.status === "new" ? "accent" : "neutral"}>{finding.status}</Badge>
        </div>
        <h2 className="mb-1.5 text-xl font-semibold tracking-tight">{finding.message}</h2>
        <div className="mt-1.5 font-mono text-xs text-muted-foreground">
          {finding.filePath} · line {finding.line}
        </div>
      </div>

      <Card className="p-4">
        <div className="mb-2 text-xs uppercase tracking-[0.08em] text-muted-foreground">Details</div>
        <dl className="grid grid-cols-2 gap-x-6 gap-y-2 text-xs">
          <dt className="text-muted-foreground">Tool</dt>
          <dd>{finding.tool}</dd>
          <dt className="text-muted-foreground">Rule</dt>
          <dd className="font-mono">{finding.ruleId}</dd>
          <dt className="text-muted-foreground">Severity</dt>
          <dd>{finding.severity}</dd>
          <dt className="text-muted-foreground">Status</dt>
          <dd>{finding.status}</dd>
          <dt className="text-muted-foreground">File</dt>
          <dd className="truncate font-mono">{finding.filePath}</dd>
          <dt className="text-muted-foreground">Line</dt>
          <dd className="font-mono">L{finding.line}</dd>
          <dt className="text-muted-foreground">Source</dt>
          <dd>{finding.source}</dd>
        </dl>
      </Card>
    </div>
  );
}

export function FindingsTab() {
  const { name } = useParams<{ name: string }>();
  const [selectedFingerprint, setSelectedFingerprint] = useState<string | null>(null);
  const [filters, setFilters] = useState<FindingFilters>({});

  const effectiveFilters: FindingFilters = { ...filters, gavelspace: name };

  const { data, isLoading } = useQuery({
    queryKey: ["gavelspace-findings", name, filters],
    queryFn: () => findingApi.listGlobalFindings(effectiveFilters),
    enabled: !!name,
  });

  const findings = data?.items ?? [];
  const selected =
    findings.find((f) => f.fingerprint === selectedFingerprint) ?? findings[0] ?? null;

  function toggleSeverity(severity: SeverityType) {
    setFilters((prev) => ({
      ...prev,
      severity: prev.severity === severity ? undefined : severity,
    }));
  }

  function clearFilters() {
    setFilters({});
  }

  const hasFilters = !!(filters.severity || filters.tool || filters.filePath);

  return (
    <div data-testid="gs-tab-findings" className="flex flex-col md:flex-row md:h-[calc(100vh-220px)]">
      <div className="flex w-full flex-col border-r border-border md:w-80 md:shrink-0">
        <div className="border-b border-border px-3.5 py-3">
          <div className="mb-2.5 flex flex-wrap gap-1.5">
            {SEVERITY_OPTIONS.map((severity) => (
              <button
                key={severity}
                onClick={() => toggleSeverity(severity)}
                className="rounded-full focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              >
                <Badge tone={filters.severity === severity ? "accent" : "neutral"}>{severity}</Badge>
              </button>
            ))}
            {hasFilters && (
              <button
                onClick={clearFilters}
                className="rounded-full focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              >
                <Badge tone="neutral">Reset</Badge>
              </button>
            )}
          </div>
          <input
            type="text"
            placeholder="Filter by file path..."
            aria-label="Filter by file path"
            className="mb-2 w-full rounded border border-border bg-card px-2 py-1 text-xs placeholder:text-muted-foreground"
            value={filters.filePath ?? ""}
            onChange={(e) =>
              setFilters((prev) => ({ ...prev, filePath: e.target.value || undefined }))
            }
          />
          <div className="text-xs text-muted-foreground" data-testid="finding-count">
            {findings.length} findings
          </div>
        </div>
        <div className="flex-1 overflow-auto">
          {isLoading ? (
            <Spinner />
          ) : findings.length === 0 ? (
            <EmptyState title="No findings" />
          ) : (
            findings.map((finding) => (
              <IssueRow
                key={finding.fingerprint}
                finding={finding}
                active={finding.fingerprint === selected?.fingerprint}
                gavelspaceName={name ?? ""}
                onClick={() => setSelectedFingerprint(finding.fingerprint)}
              />
            ))
          )}
        </div>
      </div>

      <div className="flex-1 overflow-auto">
        {selected ? (
          <IssueDetail finding={selected} />
        ) : (
          <p className="p-6 text-center text-sm text-muted-foreground">
            Select a finding to view details
          </p>
        )}
      </div>
    </div>
  );
}
