import { useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import * as casefileApi from "@/entities/casefile/api";
import { VerdictBadge } from "@/entities/casefile/ui/verdict-badge";
import { Card, CardContent } from "@/shared/ui/card";
import { Spinner } from "@/shared/ui/spinner";
import { timeAgo } from "@/shared/lib/format";

const PAGE_SIZE = 50;

export function CaseFilesTab() {
  const { name } = useParams<{ name: string }>();
  const { data, isLoading } = useQuery({
    queryKey: ["gavelspace-casefiles", name],
    queryFn: () => casefileApi.listCaseFiles({ limit: PAGE_SIZE, gavelspace: name }),
    enabled: !!name,
  });

  const casefiles = data?.items ?? [];

  return (
    <div data-testid="gs-tab-case-files">
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <Spinner />
          ) : casefiles.length === 0 ? (
            <p className="p-6 text-sm text-muted-foreground">No case files yet.</p>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border">
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">#</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Commit</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Branch</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Findings</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Verdict</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Date</th>
                </tr>
              </thead>
              <tbody>
                {casefiles.map((cf) => (
                  <tr key={cf.id} className="border-b border-border last:border-0">
                    <td className="px-4 py-3">
                      <Link
                        to={`/casefiles/${cf.id}`}
                        className="font-medium text-primary underline-offset-4 hover:underline"
                      >
                        <code className="text-xs">{cf.id.slice(0, 8)}</code>
                      </Link>
                    </td>
                    <td className="px-4 py-3">
                      {cf.commitSha ? <code className="text-xs">{cf.commitSha.slice(0, 8)}</code> : "—"}
                    </td>
                    <td className="px-4 py-3">{cf.branch || "—"}</td>
                    <td className="px-4 py-3">
                      <span>{cf.totalFindings}</span>
                      {cf.newFindings > 0 && (
                        <span className="ml-2 text-xs text-danger">+{cf.newFindings} new</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <VerdictBadge outcome={cf.verdictOutcome} />
                    </td>
                    <td className="px-4 py-3">
                      <span title={cf.createdAt}>{timeAgo(cf.createdAt)}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
