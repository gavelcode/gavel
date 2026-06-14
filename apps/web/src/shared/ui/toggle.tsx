import { cn } from "@/shared/lib/utils";

interface ToggleProps {
  checked: boolean;
  onChange?: (checked: boolean) => void;
  className?: string;
}

export function Toggle({ checked, onChange, className }: ToggleProps) {
  return (
    <button
      role="switch"
      aria-checked={checked}
      onClick={() => onChange?.(!checked)}
      className={cn(
        "relative inline-flex h-[18px] w-8 shrink-0 rounded-full border transition-colors duration-fast focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring",
        checked
          ? "border-transparent bg-primary"
          : "border-border bg-muted",
        className,
      )}
    >
      <span
        className={cn(
          "pointer-events-none absolute top-0.5 block h-3.5 w-3.5 rounded-full bg-white shadow-raised transition-transform duration-normal ease-out-expo",
          checked ? "translate-x-[14px]" : "translate-x-0.5",
        )}
      />
    </button>
  );
}
