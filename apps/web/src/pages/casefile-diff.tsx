import { useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import * as casefileApi from "@/entities/casefile/api";
import * as findingApi from "@/entities/finding/api";
import type { Finding } from "@/entities/finding/model";
import { Severity } from "@/shared/ui/severity";
import { Badge } from "@/shared/ui/badge";
import { Card, CardHeader, CardTitle, CardContent } from "@/shared/ui/card";
import { TopBar } from "@/shared/ui/top-bar";
import { Spinner } from "@/shared/ui/spinner";

function diffFindings(
  baseFindings: Finding[],
  compareFindings: Finding[],
) {
  const baseSet = new Set(baseFindings.map((f) => f.fingerprint));
  const compareSet = new Set(compareFindings.map((f) => f.fingerprint));

  const added = compareFindings.filter((f) => !baseSet.has(f.fingerprint));
  const resolved = baseFindings.filter((f) => !compareSet.has(f.fingerprint));
  const unchanged = compareFindings.filter((f) => baseSet.has(f.fingerprint));

  return { added, resolved, unchanged };
}

function FindingsList({ findings, label }: { findings: Finding[]; label: string }) {
  if (findings.length === 0) {
    return (
      <p className="py-3 text-xs text-muted-foreground">
        No {label.toLowerCase()} findings
      </p>
    );
  }

  return (
    <div className="divide-y divide-border">
      {findings.map((f) => (
        <div key={f.fingerprint} className="flex items-start justify-between py-2.5">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <Severity level={f.severity} />
              <span className="text-xs font-medium">{f.ruleId}</span>
              <span className="text-xs text-muted-foreground">{f.tool}</span>
            </div>
            <p className="mt-0.5 text-xs text-muted-foreground truncate">
              {f.message}
            </p>
            <p className="mt-0.5 font-mono text-xs text-muted-foreground">
              {f.filePath}:{f.line}
            </p>
          </div>
        </div>
      ))}
    </div>
  );
}

export function CaseFileDiffPage() {
  const { id, compareId } = useParams();

  const { data: baseCaseFile } = useQuery({
    queryKey: ["casefile", id],
    queryFn: () => casefileApi.getCaseFile(id!),
    enabled: !!id,
  });

  const { data: compareCaseFile } = useQuery({
    queryKey: ["casefile", compareId],
    queryFn: () => casefileApi.getCaseFile(compareId!),
    enabled: !!compareId,
  });

  const { data: baseFindings, isLoading: loadingBase } = useQuery({
    queryKey: ["findings", { casefileId: id }],
    queryFn: () => findingApi.listFindings(id!),
    enabled: !!id,
  });

  const { data: compareFindings, isLoading: loadingCompare } = useQuery({
    queryKey: ["findings", { casefileId: compareId }],
    queryFn: () => findingApi.listFindings(compareId!),
    enabled: !!compareId,
  });

  if (loadingBase || loadingCompare || !baseCaseFile || !compareCaseFile) {
    return <Spinner />;
  }

  const base = baseFindings?.items ?? [];
  const compare = compareFindings?.items ?? [];
  const { added, resolved, unchanged } = diffFindings(base, compare);

  return (
    <div className="flex flex-col">
      <TopBar
        crumbs={[
          "Case Files",
          `${baseCaseFile.commitSha.slice(0, 7)} vs ${compareCaseFile.commitSha.slice(0, 7)}`,
        ]}
      />
      <div className="flex-1 overflow-auto p-6">
        <div className="mb-6">
          <h1 className="font-mono text-2xl font-semibold tracking-tight">
            Run Diff
          </h1>
          <p className="mt-1 text-label text-muted-foreground">
            Comparing{" "}
            <Link to={`/casefiles/${id}`} className="text-primary underline-offset-4 hover:underline">
              <code>{baseCaseFile.commitSha.slice(0, 7)}</code>
            </Link>
            {" "}(base) with{" "}
            <Link to={`/casefiles/${compareId}`} className="text-primary underline-offset-4 hover:underline">
              <code>{compareCaseFile.commitSha.slice(0, 7)}</code>
            </Link>
            {" "}(compare)
          </p>
        </div>

        <div className="mb-6 grid grid-cols-3 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm text-muted-foreground">Added</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold text-danger">{added.length}</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm text-muted-foreground">Resolved</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold text-success">{resolved.length}</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm text-muted-foreground">Unchanged</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold">{unchanged.length}</p>
            </CardContent>
          </Card>
        </div>

        {added.length > 0 && (
          <Card className="mb-4 p-4">
            <div className="mb-3.5 flex items-center gap-2 text-label font-medium">
              Added findings
              <Badge tone="danger">{added.length}</Badge>
            </div>
            <FindingsList findings={added} label="added" />
          </Card>
        )}

        {resolved.length > 0 && (
          <Card className="mb-4 p-4">
            <div className="mb-3.5 flex items-center gap-2 text-label font-medium">
              Resolved findings
              <Badge tone="success">{resolved.length}</Badge>
            </div>
            <FindingsList findings={resolved} label="resolved" />
          </Card>
        )}

        <Card className="p-4">
          <div className="mb-3.5 flex items-center gap-2 text-label font-medium">
            Unchanged findings
            <Badge tone="neutral">{unchanged.length}</Badge>
          </div>
          <FindingsList findings={unchanged} label="unchanged" />
        </Card>
      </div>
    </div>
  );
}
