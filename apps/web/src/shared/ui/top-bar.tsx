import { useAuth } from "@/entities/user/use-auth";
import { Avatar } from "./avatar";
import { cn } from "@/shared/lib/utils";

interface TopBarProps {
  crumbs: string[];
  action?: React.ReactNode;
  className?: string;
}

export function TopBar({ crumbs, action, className }: TopBarProps) {
  const { user } = useAuth();
  const initials = user?.displayName?.slice(0, 2).toUpperCase() ?? "U";

  return (
    <header
      className={cn(
        "flex h-[52px] shrink-0 items-center gap-4 border-b border-border bg-background px-6",
        className,
      )}
    >
      <div className="flex items-center gap-2 text-label">
        {crumbs.map((c, i) => (
          <span key={i} className="flex items-center gap-2">
            {i > 0 && <span className="text-muted-foreground">/</span>}
            <span
              className={cn(
                i === crumbs.length - 1
                  ? "font-medium text-foreground"
                  : "text-muted-foreground",
              )}
            >
              {c}
            </span>
          </span>
        ))}
      </div>
      <div className="flex-1" />
      {action}
      <Avatar initials={initials} tone="violet" />
    </header>
  );
}
