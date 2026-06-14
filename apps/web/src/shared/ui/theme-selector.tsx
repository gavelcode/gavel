import { useTheme } from "@/shared/lib/use-theme";
import { Monitor, Sun, Moon } from "lucide-react";
import { cn } from "@/shared/lib/utils";

type Theme = "system" | "light" | "dark";

const OPTIONS: { value: Theme; label: string; icon: typeof Sun }[] = [
  { value: "system", label: "Auto", icon: Monitor },
  { value: "light", label: "Light", icon: Sun },
  { value: "dark", label: "Dark", icon: Moon },
];

export function ThemeSelector() {
  const { theme, setTheme, highContrast, setHighContrast } = useTheme();

  return (
    <div className="space-y-1.5">
      <fieldset
        role="radiogroup"
        aria-label="Theme"
        className="flex rounded-lg border border-border bg-muted p-0.5"
      >
        {OPTIONS.map((opt) => {
          const active = theme === opt.value;
          return (
            <label
              key={opt.value}
              className={cn(
                "flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 text-2xs font-medium transition-colors duration-fast",
                active
                  ? "bg-card text-foreground shadow-raised"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              <input
                type="radio"
                name="theme"
                value={opt.value}
                checked={active}
                onChange={() => setTheme(opt.value)}
                className="sr-only"
                aria-label={opt.label}
              />
              <opt.icon className="h-3 w-3" />
              <span>{opt.label}</span>
            </label>
          );
        })}
      </fieldset>
      <label className="flex cursor-pointer items-center gap-1.5 px-0.5 text-2xs text-muted-foreground">
        <input
          type="checkbox"
          checked={highContrast}
          onChange={(e) => setHighContrast(e.target.checked)}
          className="rounded border-border"
          aria-label="High contrast"
        />
        High contrast
      </label>
    </div>
  );
}
