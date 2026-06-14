import { useState, useMemo } from "react";
import { useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import * as casefileApi from "@/entities/casefile/api";
import * as findingApi from "@/entities/finding/api";
import { CaseFileSummaryCards } from "@/entities/casefile/ui/casefile-summary-cards";
import { VerdictBadge } from "@/entities/casefile/ui/verdict-badge";
import { FindingsTable } from "@/entities/finding/ui/findings-table";
import { Card, CardContent } from "@/shared/ui/card";
import { Input } from "@/shared/ui/input";
import { TopBar } from "@/shared/ui/top-bar";
import { ArrowLeft, GitCommit, GitBranch, Calendar } from "lucide-react";
import { timeAgo } from "@/shared/lib/format";
import type { Severity as SeverityType } from "@/entities/finding/model";

const SEVERITY_OPTIONS: SeverityType[] = ["error", "warning", "note"];

export function CaseFileDetailPage() {
  const { id } = useParams<{ id: string }>();
  const casefileId = id!;

  const [severityFilter, setSeverityFilter] = useState("");
  const [toolFilter, setToolFilter] = useState("");
  const [fileFilter, setFileFilter] = useState("");

  const { data: caseFile } = useQuery({
    queryKey: ["casefile", casefileId],
    queryFn: () => casefileApi.getCaseFile(casefileId),
  });

  const { data, isLoading } = useQuery({
    queryKey: ["findings", casefileId],
    queryFn: () => findingApi.listFindings(casefileId),
  });

  const findings = useMemo(() => data?.items ?? [], [data?.items]);

  const tools = useMemo(
    () => [...new Set(findings.map((f) => f.tool))].sort(),
    [findings],
  );

  const filteredFindings = useMemo(() => {
    let result = findings;
    if (severityFilter) result = result.filter((f) => f.severity === severityFilter);
    if (toolFilter) result = result.filter((f) => f.tool === toolFilter);
    if (fileFilter) result = result.filter((f) => f.filePath.toLowerCase().includes(fileFilter.toLowerCase()));
    return result;
  }, [findings, severityFilter, toolFilter, fileFilter]);

  return (
    <div className="flex flex-col">
      <TopBar crumbs={["Case Files", `#${id?.slice(0, 8)}`]} />
      <div className="flex flex-col gap-d-gap overflow-auto p-d-page">
      <div className="flex items-center gap-3">
        <Link to="/casefiles" className="text-muted-foreground hover:text-foreground">
          <ArrowLeft className="h-5 w-5" />
        </Link>
        {caseFile && <VerdictBadge outcome={caseFile.verdictOutcome} />}
      </div>

      {caseFile && (
        <div className="flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
          <span className="inline-flex items-center gap-1.5">
            <GitCommit className="h-4 w-4" />
            <code className="text-xs">{caseFile.commitSha.slice(0, 8)}</code>
          </span>
          <span className="inline-flex items-center gap-1.5">
            <GitBranch className="h-4 w-4" />
            {caseFile.branch}
          </span>
          <span className="inline-flex items-center gap-1.5">
            <Calendar className="h-4 w-4" />
            <span title={caseFile.createdAt}>{timeAgo(caseFile.createdAt)}</span>
          </span>
        </div>
      )}

      {caseFile && <CaseFileSummaryCards caseFile={caseFile} />}

      <div className="flex flex-wrap items-center gap-3">
        <select
          value={severityFilter}
          onChange={(e) => setSeverityFilter(e.target.value)}
          className="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
          aria-label="Filter by severity"
        >
          <option value="">All severities</option>
          {SEVERITY_OPTIONS.map((l) => (
            <option key={l} value={l}>{l}</option>
          ))}
        </select>

        <select
          value={toolFilter}
          onChange={(e) => setToolFilter(e.target.value)}
          className="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
          aria-label="Filter by tool"
        >
          <option value="">All tools</option>
          {tools.map((t) => (
            <option key={t} value={t}>{t}</option>
          ))}
        </select>

        <Input
          value={fileFilter}
          onChange={(e) => setFileFilter(e.target.value)}
          placeholder="Filter by file path..."
          className="max-w-xs"
          aria-label="Filter by file path"
        />

        <span className="text-sm text-muted-foreground">
          {filteredFindings.length} finding(s)
        </span>
      </div>

      <Card>
        <CardContent className="p-0">
          <FindingsTable findings={filteredFindings} isLoading={isLoading} />
        </CardContent>
      </Card>
      </div>
    </div>
  );
}
