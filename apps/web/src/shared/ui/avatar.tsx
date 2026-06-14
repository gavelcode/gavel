import { cn } from "@/shared/lib/utils";

const toneClasses = {
  indigo: "bg-primary",
  teal: "bg-[hsl(185_60%_45%)]",
  rose: "bg-danger",
  amber: "bg-warning",
  lime: "bg-success",
  violet: "bg-[hsl(280_60%_55%)]",
};

interface AvatarProps {
  initials: string;
  tone?: keyof typeof toneClasses;
  size?: "sm" | "md";
  className?: string;
}

export function Avatar({
  initials,
  tone = "indigo",
  size = "sm",
  className,
}: AvatarProps) {
  return (
    <div
      className={cn(
        "grid shrink-0 place-items-center rounded-full text-white",
        toneClasses[tone],
        size === "sm" ? "h-6 w-6 text-2xs" : "h-8 w-8 text-xs",
        "font-semibold tracking-wide",
        className,
      )}
    >
      {initials}
    </div>
  );
}
