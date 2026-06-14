import { Link } from "react-router-dom";
import { Card } from "@/shared/ui/card";
import { VerdictBadge } from "@/entities/casefile/ui/verdict-badge";
import { timeAgo } from "@/shared/lib/format";
import type { ProjectSummary } from "../model";

interface ProjectCardProps {
  project: ProjectSummary;
}

export function ProjectCard({ project }: ProjectCardProps) {
  return (
    <Link to={`/projects/${encodeURIComponent(project.key)}`}>
      <Card className="hover:border-primary/50 transition-colors duration-fast p-5">
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0">
            <span className="text-base font-semibold text-primary truncate">
              {project.name}
            </span>
            <p className="mt-1 text-xs text-muted-foreground">
              {project.key} &middot; {project.defaultBranch} &middot; Created {timeAgo(project.createdAt)}
            </p>
          </div>
          <div className="shrink-0">
            <VerdictBadge outcome={project.latestVerdict} />
          </div>
        </div>

        <div className="mt-4 text-sm text-muted-foreground">
          {project.totalFindings} findings
        </div>
      </Card>
    </Link>
  );
}
