import { Link } from "react-router-dom";
import { TopBar } from "@/shared/ui/top-bar";
import { Button } from "@/shared/ui/button";

export function NotFoundPage() {
  return (
    <div className="flex flex-col">
      <TopBar crumbs={["404"]} />
      <div className="flex flex-1 flex-col items-center justify-center py-24 text-center">
      <span className="text-6xl font-bold text-muted-foreground/30">404</span>
      <h1 className="mt-4 text-xl font-semibold">Page not found</h1>
      <p className="mt-2 text-sm text-muted-foreground">
        The page you're looking for doesn't exist or has been moved.
      </p>
      <Link to="/" className="mt-6">
        <Button variant="outline">Back to Projects</Button>
      </Link>
      </div>
    </div>
  );
}
