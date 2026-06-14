import { useOutletContext, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import type { GavelspaceDetail } from "@/entities/gavelspace/model";
import * as projectApi from "@/entities/project/api";
import { VerdictBadge } from "@/entities/casefile/ui/verdict-badge";
import { Card, CardContent } from "@/shared/ui/card";
import { Spinner } from "@/shared/ui/spinner";
import { timeAgo } from "@/shared/lib/format";

export function ProjectsTab() {
  const { gavelspace } = useOutletContext<{ gavelspace: GavelspaceDetail }>();

  const projectKeys = new Set(gavelspace.projects.map((p) => p.key));

  const { data, isLoading } = useQuery({
    queryKey: ["gavelspace-projects", gavelspace.name],
    queryFn: () => projectApi.listProjects({ limit: 200 }),
    enabled: projectKeys.size > 0,
  });

  const projects = (data?.items ?? []).filter((p) => projectKeys.has(p.key));

  return (
    <div data-testid="gs-tab-projects">
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <Spinner />
          ) : projects.length === 0 ? (
            <p className="p-6 text-sm text-muted-foreground">No projects yet.</p>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border">
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Project</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Verdict</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Findings</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Branch</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Created</th>
                </tr>
              </thead>
              <tbody>
                {projects.map((p) => (
                  <tr key={p.id} className="border-b border-border last:border-0">
                    <td className="px-4 py-3">
                      <Link
                        to={`/projects/${p.key}`}
                        className="font-medium text-primary underline-offset-4 hover:underline"
                      >
                        {p.name}
                      </Link>
                    </td>
                    <td className="px-4 py-3">
                      <VerdictBadge outcome={p.latestVerdict} />
                    </td>
                    <td className="px-4 py-3">{p.totalFindings}</td>
                    <td className="px-4 py-3">{p.defaultBranch}</td>
                    <td className="px-4 py-3">
                      <span title={p.createdAt}>{timeAgo(p.createdAt)}</span>
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
