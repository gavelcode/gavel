import { Link } from "react-router-dom";
import type { CaseFile } from "@/entities/casefile/model";
import { VerdictBadge } from "@/entities/casefile/ui/verdict-badge";
import { timeAgo } from "@/shared/lib/format";

interface ActivityFeedProps {
  casefiles: CaseFile[];
}

export function ActivityFeed({ casefiles }: ActivityFeedProps) {
  if (casefiles.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">No case files yet.</p>
    );
  }
  return (
    <ul className="divide-y divide-border">
      {casefiles.map((cf) => (
        <li key={cf.id}>
          <Link
            to={`/casefiles/${cf.id}`}
            data-testid="activity-feed-row"
            className="flex items-center justify-between gap-3 py-3 text-sm transition-colors duration-fast hover:bg-muted/40"
          >
            <div className="flex min-w-0 items-center gap-3">
              <VerdictBadge outcome={cf.verdictOutcome} />
              <code className="text-xs text-muted-foreground">
                {cf.commitSha ? cf.commitSha.slice(0, 7) : "—"}
              </code>
              <span className="truncate text-muted-foreground">{cf.branch}</span>
            </div>
            <span className="text-xs text-muted-foreground" title={cf.createdAt}>
              {timeAgo(cf.createdAt)}
            </span>
          </Link>
        </li>
      ))}
    </ul>
  );
}
