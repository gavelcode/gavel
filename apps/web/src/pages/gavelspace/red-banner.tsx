import { AlertOctagon } from "lucide-react";

interface RedBannerProps {
  names: string[];
}

export function RedBanner({ names }: RedBannerProps) {
  if (names.length === 0) return null;
  const noun = names.length === 1 ? "project" : "projects";
  return (
    <div
      data-testid="overview-red-banner"
      role="alert"
      className="flex items-center gap-3 rounded-lg border border-danger/40 bg-danger/10 px-4 py-3 text-sm text-danger"
    >
      <AlertOctagon className="h-4 w-4 shrink-0" />
      <span>
        <span className="font-semibold">
          {names.length} {noun} failing:
        </span>{" "}
        {names.join(", ")}
      </span>
    </div>
  );
}
