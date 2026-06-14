import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/shared/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium whitespace-nowrap border",
  {
    variants: {
      tone: {
        neutral:
          "bg-muted text-muted-foreground border-border",
        success:
          "bg-success/10 text-success border-success/30",
        warning:
          "bg-warning/10 text-warning border-warning/30",
        danger:
          "bg-danger/10 text-danger border-danger/30",
        accent:
          "bg-primary/10 text-primary border-primary/30",
      },
    },
    defaultVariants: {
      tone: "neutral",
    },
  },
);

interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {}

export function Badge({ className, tone, ...props }: BadgeProps) {
  return <span className={cn(badgeVariants({ tone }), className)} {...props} />;
}
