interface SectionLabelProps {
  children: React.ReactNode;
  collapsed?: boolean;
}

export function SectionLabel({ children, collapsed }: SectionLabelProps) {
  if (collapsed) return <div className="my-1" />;
  return (
    <div className="px-2.5 pb-1.5 pt-3 text-2xs uppercase tracking-[0.08em] text-muted-foreground/60">
      {children}
    </div>
  );
}
