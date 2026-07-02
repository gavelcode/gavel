import { useParams, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import type { Pleading } from "@/entities/pleading/model";
import * as pleadingApi from "@/entities/pleading/api";
import { Badge } from "@/shared/ui/badge";
import { Card } from "@/shared/ui/card";
import { SkeletonCard } from "@/shared/ui/skeleton";

function gateStatus(pr: Pleading): "pass" | "fail" | "none" {
  if (!pr.gateResult) return "none";
  return pr.gateResult.passed ? "pass" : "fail";
}

function GateStatusBadge({ pr }: { pr: Pleading }) {
  const status = gateStatus(pr);
  if (status === "none") return <Badge tone="neutral">Pending</Badge>;
  return status === "pass" ? (
    <Badge tone="success">Passed</Badge>
  ) : (
    <Badge tone="danger">Failed</Badge>
  );
}

export function PRChecksTab() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useQuery({
    queryKey: ["gavelspace-prs", name],
    queryFn: () => pleadingApi.listPleadings({ limit: 20, gavelspace: name }),
    enabled: !!name,
  });

  return (
    <div data-testid="gs-tab-pr-checks">
      {isLoading && (
        <div className="space-y-3">
          {Array.from({ length: 3 }, (_, i) => (
            <SkeletonCard key={i} />
          ))}
        </div>
      )}
      {data && data.items.length === 0 && (
        <p className="text-sm text-muted-foreground">No pull requests found.</p>
      )}
      <div className="space-y-3">
        {data?.items.map((pr) => (
          <Card
            key={pr.id}
            className="cursor-pointer p-4 transition-colors duration-fast hover:bg-muted/50"
            onClick={() => void navigate(`/pr-checks/${pr.id}`)}
            role="link"
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <GateStatusBadge pr={pr} />
                <span className="font-mono text-xs text-muted-foreground">
                  #{pr.number}
                </span>
                <span className="text-sm font-medium">{pr.title}</span>
              </div>
              <div className="flex items-center gap-4 text-xs text-muted-foreground">
                <span>{pr.petitioner}</span>
                <span className="rounded bg-muted px-2 py-0.5 font-mono text-foreground">
                  {pr.sourceBranch}
                </span>
                <span>→</span>
                <span className="rounded bg-muted px-2 py-0.5 font-mono text-foreground">
                  {pr.targetBranch}
                </span>
              </div>
            </div>
            {pr.gateResult && (
              <div className="mt-2 flex gap-4 text-xs text-muted-foreground">
                {pr.gateResult.conditions.map((c) => (
                  <span
                    key={c.label}
                    className={c.passed ? "text-success" : "text-danger"}
                  >
                    {c.label}: {c.value}
                  </span>
                ))}
              </div>
            )}
          </Card>
        ))}
      </div>
    </div>
  );
}
