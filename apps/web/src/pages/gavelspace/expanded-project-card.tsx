import { Link } from "react-router-dom";
import type { CaseFile } from "@/entities/casefile/model";
import { Card } from "@/shared/ui/card";
import { VerdictBadge } from "@/entities/casefile/ui/verdict-badge";
import { BeforeAfterDiff } from "./before-after-diff";

interface ExpandedProjectCardProps {
  projectName: string;
  latest: CaseFile;
  previous: CaseFile | null;
}

export function ExpandedProjectCard({ projectName, latest, previous }: ExpandedProjectCardProps) {
  const shortSha = latest.commitSha ? latest.commitSha.slice(0, 7) : "—";
  return (
    <Card className="border-danger/40 p-4">
      <div className="mb-3 flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <VerdictBadge outcome={latest.verdictOutcome} />
          <h4 className="text-sm font-semibold">{projectName}</h4>
        </div>
        <Link
          to={`/casefiles/${latest.id}`}
          className="text-xs text-primary underline-offset-4 hover:underline"
        >
          View case file
        </Link>
      </div>
      <div className="mb-2 grid grid-cols-1 gap-1 text-xs text-muted-foreground md:grid-cols-3">
        <div>
          commit <code className="text-foreground">{shortSha}</code>
        </div>
        <div>
          branch <span className="text-foreground">{latest.branch}</span>
        </div>
        <div>
          <span className="font-medium text-danger">{latest.newFindings} new findings</span>
        </div>
      </div>
      <BeforeAfterDiff before={previous?.totalFindings ?? 0} after={latest.totalFindings} />
    </Card>
  );
}
