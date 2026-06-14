import { Check, AlertTriangle, X } from "lucide-react";
import { cn } from "@/shared/lib/utils";

interface Stat {
  label: string;
  value: string;
  tone?: "success" | "warning" | "danger" | "neutral";
}

interface QualityGateBannerProps {
  status: "pass" | "warn" | "fail";
  title: string;
  description: string;
  stats?: Stat[];
  className?: string;
}

const statusConfig = {
  pass: {
    icon: Check,
    border: "border-success/30",
    bg: "bg-success/5",
    iconBg: "bg-success/20",
    iconColor: "text-success",
  },
  warn: {
    icon: AlertTriangle,
    border: "border-warning/30",
    bg: "bg-warning/5",
    iconBg: "bg-warning/20",
    iconColor: "text-warning",
  },
  fail: {
    icon: X,
    border: "border-danger/30",
    bg: "bg-danger/5",
    iconBg: "bg-danger/20",
    iconColor: "text-danger",
  },
};

export function QualityGateBanner({
  status,
  title,
  description,
  stats,
  className,
}: QualityGateBannerProps) {
  const config = statusConfig[status];
  const Icon = config.icon;

  return (
    <div
      className={cn(
        "flex items-center gap-4 rounded-xl border p-4",
        config.border,
        config.bg,
        className,
      )}
    >
      <div
        className={cn(
          "grid h-11 w-11 shrink-0 place-items-center rounded-lg",
          config.iconBg,
          config.iconColor,
        )}
      >
        <Icon className="h-5 w-5" />
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-sm font-medium">{title}</div>
        <div className="mt-0.5 text-xs text-muted-foreground">
          {description}
        </div>
      </div>
      {stats && (
        <div className="flex gap-6">
          {stats.map((s) => (
            <div key={s.label} className="text-right">
              <div
                className={cn(
                  "font-mono text-lg font-medium",
                  s.tone === "success" && "text-success",
                  s.tone === "warning" && "text-warning",
                  s.tone === "danger" && "text-danger",
                  (!s.tone || s.tone === "neutral") && "text-foreground",
                )}
              >
                {s.value}
              </div>
              <div className="text-2xs uppercase tracking-wider text-muted-foreground">
                {s.label}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
