import { useRef } from "react";
import { cn } from "@/shared/lib/utils";

interface TabsProps {
  items: string[];
  active: number;
  onChange?: (index: number) => void;
  className?: string;
  variant?: "underline" | "pill";
}

export function Tabs({ items, active, onChange, className, variant = "underline" }: TabsProps) {
  const isUnderline = variant === "underline";
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([]);

  function handleKeyDown(e: React.KeyboardEvent, index: number) {
    let next: number;
    if (e.key === "ArrowRight") {
      e.preventDefault();
      next = (index + 1) % items.length;
    } else if (e.key === "ArrowLeft") {
      e.preventDefault();
      next = (index - 1 + items.length) % items.length;
    } else if (e.key === "Home") {
      e.preventDefault();
      next = 0;
    } else if (e.key === "End") {
      e.preventDefault();
      next = items.length - 1;
    } else {
      return;
    }
    onChange?.(next);
    tabRefs.current[next]?.focus();
  }

  return (
    <div
      role="tablist"
      className={cn(
        "flex gap-1",
        isUnderline && "gap-6 border-b border-border",
        className,
      )}
    >
      {items.map((label, i) => (
        <button
          key={label}
          ref={(el) => { tabRefs.current[i] = el; }}
          role="tab"
          aria-selected={i === active}
          tabIndex={i === active ? 0 : -1}
          onClick={() => onChange?.(i)}
          onKeyDown={(e) => handleKeyDown(e, i)}
          className={cn(
            "text-sm transition-colors duration-fast focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring",
            isUnderline && cn(
              "-mb-px border-b-2 pb-2.5 pt-2.5",
              i === active
                ? "border-primary font-medium text-foreground"
                : "border-transparent text-muted-foreground hover:text-foreground",
            ),
            !isUnderline && cn(
              "rounded-md px-3 py-1.5",
              i === active
                ? "bg-muted font-medium text-foreground"
                : "text-muted-foreground hover:bg-muted/50 hover:text-foreground",
            ),
          )}
        >
          {label}
        </button>
      ))}
    </div>
  );
}
