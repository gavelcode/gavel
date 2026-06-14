import { useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import * as projectApi from "@/entities/project/api";
import * as casefileApi from "@/entities/casefile/api";
import * as findingApi from "@/entities/finding/api";
import { VerdictBadge } from "@/entities/casefile/ui/verdict-badge";
import { Severity } from "@/shared/ui/severity";
import { TopBar } from "@/shared/ui/top-bar";
import { Badge } from "@/shared/ui/badge";
import { Card } from "@/shared/ui/card";
import { Skeleton, SkeletonText } from "@/shared/ui/skeleton";
import { timeAgo } from "@/shared/lib/format";
import { GitBranch } from "lucide-react";
import { Button } from "@/shared/ui/button";
import {
  BarChart,
  Bar,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";

export function ProjectDetailPage() {
  const { key } = useParams();

  const { data: project, isLoading } = useQuery({
    queryKey: ["project", key],
    queryFn: () => projectApi.getProject(key!),
    enabled: !!key,
  });

  const { data: casefilesData } = useQuery({
    queryKey: ["casefiles", { projectId: key }],
    queryFn: () => casefileApi.listCaseFiles({ limit: 5, projectId: project!.id }),
    enabled: !!project,
  });

  const { data: trendData } = useQuery({
    queryKey: ["casefiles-trend", { projectId: key }],
    queryFn: () => casefileApi.listCaseFiles({ limit: 20, projectId: project!.id }),
    enabled: !!project,
  });

  const { data: findingsData } = useQuery({
    queryKey: ["findings", { projectId: key }],
    queryFn: () => findingApi.listGlobalFindings({ projectId: project!.id }),
    enabled: !!project,
  });

  if (isLoading || !project) {
    return (
      <div className="flex flex-col">
        <div className="border-b border-border px-6 py-3">
          <Skeleton className="h-5 w-48" />
        </div>
        <div className="flex-1 overflow-auto p-6">
          <div className="mb-6">
            <Skeleton className="mb-2 h-8 w-64" />
            <Skeleton className="h-4 w-40" />
          </div>
          <div className="mb-4 grid grid-cols-1 gap-4 lg:grid-cols-2">
            <div className="rounded-xl border border-border p-4">
              <Skeleton className="mb-4 h-5 w-32" />
              <Skeleton className="h-[180px] w-full" />
            </div>
            <div className="rounded-xl border border-border p-4">
              <Skeleton className="mb-4 h-5 w-24" />
              <SkeletonText lines={4} />
            </div>
          </div>
        </div>
      </div>
    );
  }

  const casefiles = casefilesData?.items ?? [];
  const findings = findingsData?.items ?? [];
  const trendCasefiles = trendData?.items ?? [];

  const coverageTrend = [...trendCasefiles]
    .filter((cf) => cf.coveragePercent !== null)
    .reverse()
    .map((cf) => ({
      commit: cf.commitSha.slice(0, 7),
      coverage: Math.round((cf.coveragePercent as number) * 10) / 10,
    }));

  const severityCounts = project.severityCounts ?? {};
  const chartData = [
    { name: "Error", value: severityCounts.error ?? 0, fill: "hsl(var(--danger))" },
    { name: "Warning", value: severityCounts.warning ?? 0, fill: "hsl(var(--warning))" },
    { name: "Note", value: severityCounts.note ?? 0, fill: "hsl(var(--muted-foreground))" },
  ];

  return (
    <div className="flex flex-col">
      <TopBar
        crumbs={[project.key, project.name]}
        action={
          <Button variant="ghost" size="sm" className="gap-1.5 font-mono text-xs">
            <GitBranch className="h-3 w-3" />
            {project.defaultBranch}
          </Button>
        }
      />
      <div className="flex-1 overflow-auto p-6">
        <div className="mb-6">
          <div className="mb-1.5 flex items-center gap-2.5">
            <h1 className="font-mono text-2xl font-semibold tracking-tight">
              {project.name}
            </h1>
            <VerdictBadge outcome={project.latestVerdict} />
          </div>
          <p className="max-w-xl text-label text-muted-foreground">
            {project.key} &middot; Created {timeAgo(project.createdAt)}
          </p>
          <p className="mt-1 max-w-xl text-label text-muted-foreground">
            {project.totalFindings} findings
          </p>
        </div>

        <div className="mb-4 grid grid-cols-1 gap-4 lg:grid-cols-2">
          <Card className="p-4">
            <div className="mb-3.5 text-label font-medium">
              Findings by severity
            </div>
            <ResponsiveContainer width="100%" height={180}>
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                <XAxis dataKey="name" tick={{ fontSize: 10 }} className="fill-muted-foreground" />
                <YAxis tick={{ fontSize: 10 }} className="fill-muted-foreground" />
                <Tooltip
                  contentStyle={{
                    background: "hsl(var(--card))",
                    border: "1px solid hsl(var(--border))",
                    borderRadius: 8,
                    fontSize: 12,
                    color: "hsl(var(--foreground))",
                  }}
                />
                <Bar dataKey="value" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
            <div className="mt-3.5 flex justify-between border-t border-border pt-3.5 text-xs text-muted-foreground">
              <span>{project.totalFindings} total findings</span>
            </div>
          </Card>

          <Card className="p-4">
            <div className="mb-3.5 text-label font-medium">Project info</div>
            <dl className="space-y-2 text-xs">
              {project.targetPattern && (
                <>
                  <dt className="text-muted-foreground">Target pattern</dt>
                  <dd className="font-mono">{project.targetPattern}</dd>
                </>
              )}
              {project.languages.length > 0 && (
                <>
                  <dt className="text-muted-foreground">Languages</dt>
                  <dd className="flex flex-wrap gap-1.5">
                    {project.languages.map((lang) => (
                      <Badge key={lang} tone="neutral">{lang}</Badge>
                    ))}
                  </dd>
                </>
              )}
            </dl>

            {project.qualityGateRules.length > 0 && (
              <div className="mt-4">
                <div className="mb-2 text-label font-medium">Quality gate rules</div>
                <div className="space-y-1.5">
                  {project.qualityGateRules.map((rule) => (
                    <div key={rule.subtype} className="flex items-center gap-2 text-xs">
                      <span>{rule.subtype}</span>
                      <span className="text-muted-foreground">({rule.strategyType})</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </Card>
        </div>

        {coverageTrend.length > 0 && (
          <Card className="mb-4 p-4">
            <div className="mb-3.5 text-label font-medium">
              Coverage trend
            </div>
            <ResponsiveContainer width="100%" height={180}>
              <AreaChart data={coverageTrend}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                <XAxis dataKey="commit" tick={{ fontSize: 10 }} className="fill-muted-foreground" />
                <YAxis domain={[0, 100]} tick={{ fontSize: 10 }} className="fill-muted-foreground" unit="%" />
                <Tooltip
                  contentStyle={{
                    background: "hsl(var(--card))",
                    border: "1px solid hsl(var(--border))",
                    borderRadius: 8,
                    fontSize: 12,
                    color: "hsl(var(--foreground))",
                  }}
                  formatter={(value) => [`${value}%`, "Coverage"]}
                />
                <Area
                  type="monotone"
                  dataKey="coverage"
                  stroke="hsl(var(--success))"
                  fill="hsl(var(--success))"
                  fillOpacity={0.15}
                  strokeWidth={2}
                />
              </AreaChart>
            </ResponsiveContainer>
          </Card>
        )}

        <Card className="mb-4 p-4">
          <div className="mb-3.5 text-label font-medium">Recent Case Files</div>
          {casefiles.length === 0 ? (
            <p className="text-xs text-muted-foreground">No case files yet</p>
          ) : (
            <div className="divide-y divide-border">
              {casefiles.map((cf) => (
                <Link
                  key={cf.id}
                  to={`/casefiles/${cf.id}`}
                  className="flex items-center justify-between py-2.5 hover:bg-muted/30 -mx-2 px-2 rounded transition-colors duration-fast"
                >
                  <div className="min-w-0">
                    <span className="font-mono text-xs">{cf.commitSha.slice(0, 7)}</span>
                    <span className="ml-2 text-xs text-muted-foreground">{cf.branch}</span>
                    <span className="ml-2 text-xs text-muted-foreground">{timeAgo(cf.createdAt)}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-xs text-muted-foreground">
                      {cf.totalFindings} findings
                    </span>
                    <VerdictBadge outcome={cf.verdictOutcome} />
                  </div>
                </Link>
              ))}
            </div>
          )}
        </Card>

        <Card className="p-4">
          <div className="mb-3.5 text-label font-medium">
            Findings ({findings.length})
          </div>
          {findings.length === 0 ? (
            <p className="text-xs text-muted-foreground">No findings</p>
          ) : (
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
                  <Badge tone={f.status === "new" ? "danger" : f.status === "resolved" ? "success" : "warning"}>
                    {f.status}
                  </Badge>
                </div>
              ))}
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}
