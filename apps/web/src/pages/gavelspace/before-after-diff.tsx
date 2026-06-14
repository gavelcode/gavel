import { cn } from "@/shared/lib/utils";

interface BeforeAfterDiffProps {
  before: number;
  after: number;
}

export function BeforeAfterDiff({ before, after }: BeforeAfterDiffProps) {
  const delta = after - before;
  return (
    <div
      data-testid="overview-before-after-diff"
      className="flex items-center gap-2 text-xs"
    >
      <span className="text-muted-foreground">before: {before}</span>
      <span className="text-muted-foreground">→</span>
      <span
        className={cn(
          "font-medium",
          delta > 0 ? "text-danger" : delta < 0 ? "text-success" : "text-muted-foreground",
        )}
      >
        after: {after}
      </span>
    </div>
  );
}
