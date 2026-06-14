import { NavLink } from "react-router-dom";
import { cn } from "@/shared/lib/utils";

interface NavItemProps {
  to: string;
  icon: React.ReactNode;
  label: string;
  badge?: string | number;
  collapsed?: boolean;
  onClick?: () => void;
}

export function NavItem({ to, icon, label, badge, collapsed, onClick }: NavItemProps) {
  return (
    <NavLink
      to={to}
      onClick={onClick}
      title={collapsed ? label : undefined}
      className={({ isActive }) =>
        cn(
          "flex items-center rounded-lg text-label transition-colors duration-fast border",
          collapsed ? "justify-center px-0 py-[7px]" : "gap-2.5 px-2.5 py-[7px]",
          isActive
            ? "border-border bg-muted font-medium text-foreground"
            : "border-transparent text-muted-foreground hover:bg-muted/50 hover:text-foreground",
        )
      }
    >
      {({ isActive }) => (
        <>
          <span
            className={cn(
              "h-3.5 w-3.5 shrink-0",
              isActive ? "text-primary" : "text-muted-foreground/60",
            )}
          >
            {icon}
          </span>
          {!collapsed && <span className="flex-1">{label}</span>}
          {!collapsed && badge != null && (
            <span className="font-mono text-2xs text-muted-foreground/60">
              {badge}
            </span>
          )}
        </>
      )}
    </NavLink>
  );
}
