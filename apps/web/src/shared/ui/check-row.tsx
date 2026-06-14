import { Check, X } from "lucide-react";
import { cn } from "@/shared/lib/utils";

interface CheckRowProps {
  label: string;
  status: "pass" | "fail" | "warn";
  value: string;
}

export function CheckRow({ label, status, value }: CheckRowProps) {
  return (
    <div className="flex items-center gap-3 border-b border-border py-2.5">
      <div
        className={cn(
          "grid h-[18px] w-[18px] place-items-center rounded-full",
          status === "pass" && "bg-success/20 text-success",
          status === "fail" && "bg-danger/20 text-danger",
          status === "warn" && "bg-warning/20 text-warning",
        )}
      >
        {status === "pass" ? (
          <Check className="h-2.5 w-2.5" />
        ) : (
          <X className="h-2.5 w-2.5" />
        )}
      </div>
      <span className="flex-1 text-xs">{label}</span>
      <span className="font-mono text-xs text-muted-foreground">
        {value}
      </span>
    </div>
  );
}
