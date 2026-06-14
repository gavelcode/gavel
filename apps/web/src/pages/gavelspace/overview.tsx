import { useMemo } from "react";
import { useOutletContext } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import type { GavelspaceDetail } from "@/entities/gavelspace/model";
import type { CaseFile } from "@/entities/casefile/model";
import * as casefileApi from "@/entities/casefile/api";
import { VerdictBadge } from "@/entities/casefile/ui/verdict-badge";
import { Card } from "@/shared/ui/card";
import { Spinner } from "@/shared/ui/spinner";
import { timeAgo } from "@/shared/lib/format";
import { ProjectStrip } from "./project-strip";
import { Sparkline } from "./sparkline";
import { ActivityFeed } from "./activity-feed";
import { RedBanner } from "./red-banner";
import { ExpandedProjectCard } from "./expanded-project-card";

const ACTIVITY_LIMIT = 5;
const HISTORY_DAYS = 7;
const HISTORY_FETCH_LIMIT = 50;

function aggregateGavelspaceVerdict(casefiles: CaseFile[]): string {
  if (casefiles.length === 0) return "";
  if (casefiles.some((cf) => cf.verdictOutcome === "fail")) return "fail";
  return "pass";
}

function groupLatestByProject(casefiles: CaseFile[]): Map<string, { latest: CaseFile; previous: CaseFile | null }> {
  const byProject = new Map<string, CaseFile[]>();
  for (const cf of casefiles) {
    const list = byProject.get(cf.projectId) ?? [];
    list.push(cf);
    byProject.set(cf.projectId, list);
  }
  const out = new Map<string, { latest: CaseFile; previous: CaseFile | null }>();
  for (const [projectId, list] of byProject) {
    out.set(projectId, { latest: list[0], previous: list[1] ?? null });
  }
  return out;
}

function bucketByDay(casefiles: CaseFile[], days: number) {
  const today = new Date();
  today.setUTCHours(0, 0, 0, 0);
  const buckets = new Map<string, number>();
  for (let i = days - 1; i >= 0; i--) {
    const d = new Date(today);
    d.setUTCDate(d.getUTCDate() - i);
    buckets.set(d.toISOString().slice(0, 10), 0);
  }
  for (const cf of casefiles) {
    const dayKey = cf.createdAt.slice(0, 10);
    if (buckets.has(dayKey)) {
      buckets.set(dayKey, (buckets.get(dayKey) ?? 0) + cf.totalFindings);
    }
  }
  return Array.from(buckets, ([day, count]) => ({ day, count }));
}

export function OverviewTab() {
  const { gavelspace } = useOutletContext<{ gavelspace: GavelspaceDetail }>();

  const { data, isLoading } = useQuery({
    queryKey: ["gavelspace-overview-casefiles", gavelspace.name],
    queryFn: () =>
      casefileApi.listCaseFiles({ limit: HISTORY_FETCH_LIMIT, gavelspace: gavelspace.name }),
  });

  const casefiles = useMemo(() => data?.items ?? [], [data?.items]);

  const verdict = useMemo(
    () => aggregateGavelspaceVerdict(casefiles),
    [casefiles],
  );
  const series = useMemo(() => bucketByDay(casefiles, HISTORY_DAYS), [casefiles]);
  const recent = casefiles.slice(0, ACTIVITY_LIMIT);
  const lastActivityAt = casefiles[0]?.createdAt;

  const latestByProject = useMemo(() => groupLatestByProject(casefiles), [casefiles]);
  const failingProjects = useMemo(
    () =>
      gavelspace.projects.filter(
        (p) => latestByProject.get(p.id)?.latest.verdictOutcome === "fail",
      ),
    [gavelspace.projects, latestByProject],
  );

  if (isLoading) {
    return <Spinner />;
  }

  if (casefiles.length === 0) {
    return (
      <div data-testid="gs-tab-overview" className="space-y-6">
        <Card className="p-6 text-center">
          <h3 className="mb-1 text-base font-semibold">No case files yet</h3>
          <p className="text-sm text-muted-foreground">
            Run <code className="rounded bg-muted px-1.5 py-0.5 font-mono">gavel judge --server</code> to record your first analysis.
          </p>
        </Card>
        <ProjectStrip projects={gavelspace.projects} />
      </div>
    );
  }

  return (
    <div data-testid="gs-tab-overview" className="space-y-6">
      {failingProjects.length > 0 && (
        <RedBanner names={failingProjects.map((p) => p.name)} />
      )}

      <div className="flex items-center justify-between gap-3">
        <div data-testid="overview-verdict-pill">
          <VerdictBadge outcome={verdict} />
        </div>
        <div className="flex items-center gap-3">
          <Sparkline series={series} />
          {lastActivityAt && (
            <span data-testid="overview-last-activity" className="text-xs text-muted-foreground" title={lastActivityAt}>
              Last activity {timeAgo(lastActivityAt)}
            </span>
          )}
        </div>
      </div>

      {failingProjects.length > 0 && (
        <div className="space-y-3">
          {failingProjects.map((p) => {
            const entry = latestByProject.get(p.id);
            if (!entry) return null;
            return (
              <ExpandedProjectCard
                key={p.id}
                projectName={p.name}
                latest={entry.latest}
                previous={entry.previous}
              />
            );
          })}
        </div>
      )}

      <ProjectStrip projects={gavelspace.projects} />

      <div>
        <h3 className="mb-2 text-xs font-medium uppercase tracking-wider text-muted-foreground">
          Recent activity
        </h3>
        <ActivityFeed casefiles={recent} />
      </div>
    </div>
  );
}
