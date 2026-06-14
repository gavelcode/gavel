import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import * as gsApi from "@/entities/gavelspace/api";
import { TopBar } from "@/shared/ui/top-bar";
import { Badge } from "@/shared/ui/badge";
import { Card } from "@/shared/ui/card";
import { SkeletonCard } from "@/shared/ui/skeleton";
import { EmptyState } from "@/shared/ui/empty-state";

export function GavelspacesPage() {
  const navigate = useNavigate();
  const { data, isLoading } = useQuery({
    queryKey: ["gavelspaces"],
    queryFn: () => gsApi.listGavelspaces(),
  });

  return (
    <div className="flex flex-col">
      <TopBar crumbs={["Gavelspaces"]} />
      <div className="flex-1 overflow-auto p-6">
        {isLoading && (
          <div className="space-y-3">
            {Array.from({ length: 3 }, (_, i) => (
              <SkeletonCard key={i} />
            ))}
          </div>
        )}
        {data && data.items.length === 0 && (
          <EmptyState title="No gavelspaces found" />
        )}
        <div className="space-y-3">
          {data?.items.map((gs) => (
            <Card
              key={gs.name}
              className="cursor-pointer p-4 transition-colors duration-fast hover:bg-muted/50"
              onClick={() => navigate(`/gavelspaces/${gs.name}`)}
              role="link"
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <span className="text-sm font-medium">{gs.name}</span>
                </div>
                <div className="flex items-center gap-4 text-xs text-muted-foreground">
                  <Badge tone="neutral">
                    {gs.projectCount} {gs.projectCount === 1 ? "project" : "projects"}
                  </Badge>
                  <span>Created {new Date(gs.createdAt).toLocaleDateString()}</span>
                </div>
              </div>
            </Card>
          ))}
        </div>
      </div>
    </div>
  );
}
