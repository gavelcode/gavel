import { useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import * as pleadingApi from "@/entities/pleading/api";
import { TopBar } from "@/shared/ui/top-bar";
import { QualityGateBanner } from "@/shared/ui/quality-gate-banner";
import { Badge } from "@/shared/ui/badge";
import { Avatar } from "@/shared/ui/avatar";
import { Card } from "@/shared/ui/card";
import { CheckRow } from "@/shared/ui/check-row";
import { Skeleton, SkeletonText } from "@/shared/ui/skeleton";

export function PRCheckDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { data: pr, isLoading } = useQuery({
    queryKey: ["pull-request", id],
    queryFn: () => pleadingApi.getPleading(id!),
    enabled: !!id,
  });

  if (isLoading || !pr) {
    return (
      <div className="flex flex-col">
        <TopBar crumbs={["PR Checks", "..."]} />
        <div className="flex-1 overflow-auto p-6">
          <div className="mb-5">
            <Skeleton className="mb-2 h-6 w-24" />
            <Skeleton className="mb-2 h-8 w-80" />
            <Skeleton className="h-4 w-64" />
          </div>
          <Skeleton className="mb-4 h-16 w-full rounded-xl" />
          <div className="rounded-xl border border-border p-4">
            <SkeletonText lines={3} />
          </div>
        </div>
      </div>
    );
  }

  const hasGate = !!pr.gateResult;
  const gateStatus = hasGate && pr.gateResult!.passed ? "pass" : "fail";
  const gateTitle = !hasGate
    ? "No quality gate result"
    : pr.gateResult!.passed
      ? "Quality gate passed"
      : "Quality gate failed";

  const conditions = pr.gateResult?.conditions ?? [];
  const conditionCount = conditions.length;
  const passedCount = conditions.filter((c) => c.passed).length;
  const gateDescription = conditionCount > 0
    ? `${passedCount} of ${conditionCount} conditions met.`
    : "No conditions evaluated.";

  return (
    <div className="flex flex-col">
      <TopBar crumbs={["PR Checks", `#${pr.number}`]} />
      <div className="flex-1 overflow-auto p-6">
        <div className="mb-5 flex items-start gap-5">
          <div className="flex-1">
            <div className="mb-2 flex items-center gap-2.5">
              <Badge tone="accent">● {pr.status}</Badge>
              <span className="font-mono text-xs text-muted-foreground">
                #{pr.number}
              </span>
            </div>
            <h1 className="mb-2 text-2xl font-semibold tracking-tight">
              {pr.title}
            </h1>
            <div className="flex items-center gap-2.5 text-xs text-muted-foreground">
              <Avatar
                initials={pr.petitioner
                  .split(" ")
                  .map((w) => w[0])
                  .join("")
                  .slice(0, 2)
                  .toUpperCase()}
                tone="teal"
              />
              <span>{pr.petitioner}</span>
              <span className="text-muted-foreground/50">wants to merge</span>
              <span className="rounded bg-muted px-2 py-0.5 font-mono text-foreground">
                {pr.sourceBranch}
              </span>
              <span className="text-muted-foreground/50">into</span>
              <span className="rounded bg-muted px-2 py-0.5 font-mono text-foreground">
                {pr.targetBranch}
              </span>
            </div>
          </div>
        </div>

        <QualityGateBanner
          status={gateStatus}
          title={gateTitle}
          description={gateDescription}
          className="mb-4"
        />

        {conditions.length > 0 && (
          <Card className="p-4">
            <div className="mb-1.5 text-label font-medium">
              Quality gate conditions
            </div>
            <div>
              {conditions.map((c) => (
                <CheckRow
                  key={c.label}
                  label={c.label}
                  status={c.passed ? "pass" : "fail"}
                  value={`${c.value} ${c.operator} ${c.threshold}`}
                />
              ))}
            </div>
            <div className="mt-3.5 flex items-center gap-2.5 border-t border-border pt-3.5 text-xs text-muted-foreground">
              <Avatar initials="GV" tone="indigo" />
              <span>
                Commit{" "}
                <span className="font-mono text-foreground">
                  {pr.commitSha.slice(0, 7)}
                </span>
              </span>
            </div>
          </Card>
        )}
      </div>
    </div>
  );
}
