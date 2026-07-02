import { useNavigate } from "react-router-dom";
import type { ProjectRef } from "@/entities/gavelspace/model";
import { VerdictBadge } from "@/entities/casefile/ui/verdict-badge";
import { Card } from "@/shared/ui/card";
import { cn } from "@/shared/lib/utils";

interface ProjectStripProps {
  projects: ProjectRef[];
}

export function ProjectStrip({ projects }: ProjectStripProps) {
  const navigate = useNavigate();
  if (projects.length === 0) {
    return null;
  }

  return (
    <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-3">
      {projects.map((p) => {
        const isCalm = p.latestVerdict === "pass" || p.latestVerdict === "";
        return (
          <Card
            key={p.id}
            data-testid="project-strip-card"
            className={cn(
              "cursor-pointer p-3 transition-colors duration-fast hover:bg-muted/50",
              isCalm ? "calm border-muted" : "border-danger/40",
            )}
            onClick={() => void navigate(`/projects/${encodeURIComponent(p.key)}`)}
            role="link"
          >
            <div className="flex items-center justify-between gap-2">
              <span className="truncate text-sm font-medium">{p.name}</span>
              <VerdictBadge outcome={p.latestVerdict} />
            </div>
            <div className="mt-1 font-mono text-2xs text-muted-foreground">{p.key}</div>
          </Card>
        );
      })}
    </div>
  );
}
