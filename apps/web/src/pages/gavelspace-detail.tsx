import { useParams, NavLink, Outlet } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import * as gsApi from "@/entities/gavelspace/api";
import { TopBar } from "@/shared/ui/top-bar";
import { Badge } from "@/shared/ui/badge";
import { Spinner } from "@/shared/ui/spinner";
import { cn } from "@/shared/lib/utils";

const tabs = [
  { to: "", label: "Overview", end: true },
  { to: "projects", label: "Projects" },
  { to: "findings", label: "Findings" },
  { to: "pr-checks", label: "PR Checks" },
  { to: "case-files", label: "Case Files" },
] as const;

export function GavelspaceDetailPage() {
  const { name } = useParams<{ name: string }>();
  const { data, isLoading, error } = useQuery({
    queryKey: ["gavelspace", name],
    queryFn: () => gsApi.getGavelspace(name!),
    enabled: !!name,
  });

  if (isLoading) {
    return (
      <div className="flex flex-col">
        <TopBar crumbs={["Gavelspaces", name ?? ""]} />
        <div className="flex-1 p-6"><Spinner /></div>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="flex flex-col">
        <TopBar crumbs={["Gavelspaces", name ?? ""]} />
        <div className="flex-1 p-6">
          <p className="text-sm text-muted-foreground">Gavelspace not found.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col">
      <TopBar crumbs={["Gavelspaces", data.name]} />
      <div className="flex-1 overflow-auto">
        <div className="px-6 pt-6">
          <div className="mb-4 flex items-center gap-3">
            <h2 className="text-lg font-semibold">{data.name}</h2>
            <Badge tone="neutral">
              {data.projects.length} {data.projects.length === 1 ? "project" : "projects"}
            </Badge>
            <span className="text-xs text-muted-foreground">
              Created {new Date(data.createdAt).toLocaleDateString()}
            </span>
          </div>

          <nav aria-label="Gavelspace sections" className="flex gap-1 border-b border-border">
            {tabs.map((tab) => (
              <NavLink
                key={tab.label}
                to={tab.to}
                end={"end" in tab ? tab.end : false}
                className={({ isActive }) =>
                  cn(
                    "px-3 py-2 text-sm border-b-2 -mb-px transition-colors duration-fast",
                    isActive
                      ? "border-primary text-foreground"
                      : "border-transparent text-muted-foreground hover:text-foreground",
                  )
                }
              >
                {tab.label}
              </NavLink>
            ))}
          </nav>
        </div>

        <div className="p-6">
          <Outlet context={{ gavelspace: data }} />
        </div>
      </div>
    </div>
  );
}
