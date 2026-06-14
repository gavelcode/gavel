import { CheckCircle2, XCircle, MinusCircle } from "lucide-react";

interface VerdictBadgeProps {
  outcome: string;
}

export function VerdictBadge({ outcome }: VerdictBadgeProps) {
  if (outcome === "pass") {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-success/15 px-2.5 py-0.5 text-xs font-medium text-success">
        <CheckCircle2 className="h-3.5 w-3.5" />
        Passed
      </span>
    );
  }


  if (outcome === "fail") {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-danger/15 px-2.5 py-0.5 text-xs font-medium text-danger">
        <XCircle className="h-3.5 w-3.5" />
        Failed
      </span>
    );
  }

  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-muted px-2.5 py-0.5 text-xs font-medium text-muted-foreground">
      <MinusCircle className="h-3.5 w-3.5" />
      Pending
    </span>
  );
}
