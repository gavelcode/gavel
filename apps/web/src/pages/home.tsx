import { Navigate, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import * as gsApi from "@/entities/gavelspace/api";
import { TopBar } from "@/shared/ui/top-bar";
import { Card } from "@/shared/ui/card";
import { Badge } from "@/shared/ui/badge";
import { Spinner } from "@/shared/ui/spinner";

export function HomePage() {
  const navigate = useNavigate();
  const { data, isLoading } = useQuery({
    queryKey: ["gavelspaces"],
    queryFn: () => gsApi.listGavelspaces(),
  });

  if (isLoading) {
    return (
      <div className="flex flex-col">
        <TopBar crumbs={["Home"]} />
        <div className="flex-1 p-6"><Spinner /></div>
      </div>
    );
  }

  const items = data?.items ?? [];

  if (items.length === 1) {
    return <Navigate to={`/gavelspaces/${encodeURIComponent(items[0].name)}`} replace />;
  }

  if (items.length === 0) {
    return (
      <div className="flex flex-col">
        <TopBar crumbs={["Home"]} />
        <div className="flex-1 p-6">
          <Card className="p-6 text-center">
            <h2 className="mb-2 text-base font-semibold">No gavelspaces yet</h2>
            <p className="text-sm text-muted-foreground">
              Run <code className="rounded bg-muted px-1.5 py-0.5 font-mono">gavel judge --server</code> in your repo to register one.
            </p>
          </Card>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col">
      <TopBar crumbs={["Home"]} />
      <div className="flex-1 overflow-auto p-6">
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {items.map((gs) => (
            <Card
              key={gs.name}
              className="cursor-pointer p-5 transition-colors duration-fast hover:bg-muted/50"
              onClick={() => navigate(`/gavelspaces/${encodeURIComponent(gs.name)}`)}
              role="link"
            >
              <div className="mb-3 flex items-center justify-between">
                <h3 className="text-base font-semibold tracking-tight">{gs.name}</h3>
                <Badge tone="neutral">
                  {gs.projectCount} {gs.projectCount === 1 ? "project" : "projects"}
                </Badge>
              </div>
              <div className="text-xs text-muted-foreground">
                Created {new Date(gs.createdAt).toLocaleDateString()}
              </div>
            </Card>
          ))}
        </div>
      </div>
    </div>
  );
}
