import { cn } from "@/shared/lib/utils";

const levels: Record<string, { color: string; label: string }> = {
  error: { color: "bg-danger", label: "Error" },
  warning: { color: "bg-warning", label: "Warning" },
  note: { color: "bg-muted-foreground/50", label: "Note" },
};

const fallback = { color: "bg-muted-foreground/50", label: "Unknown" };

interface SeverityProps {
  level: string;
  className?: string;
}

export function Severity({ level, className }: SeverityProps) {
  const { color, label } = levels[level] ?? fallback;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 text-xs text-muted-foreground",
        className,
      )}
    >
      <span
        className={cn("inline-block h-2 w-2 rounded-full", color)}
        role="img"
        aria-label={label}
      />
      {label}
    </span>
  );
}
