import { TopBar } from "@/shared/ui/top-bar";
import { Button } from "@/shared/ui/button";
import { Badge } from "@/shared/ui/badge";
import { Card, CardHeader, CardTitle, CardContent } from "@/shared/ui/card";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { Progress } from "@/shared/ui/progress";
import { Severity } from "@/shared/ui/severity";
import { Tabs } from "@/shared/ui/tabs";
import { Toggle } from "@/shared/ui/toggle";
import { Skeleton, SkeletonText, SkeletonCard } from "@/shared/ui/skeleton";
import { EmptyState } from "@/shared/ui/empty-state";
import { FolderKanban, Search } from "lucide-react";
import { useState } from "react";

const COLORS = [
  { name: "Background", var: "--background", class: "bg-background" },
  { name: "Foreground", var: "--foreground", class: "bg-foreground" },
  { name: "Primary", var: "--primary", class: "bg-primary" },
  { name: "Secondary", var: "--secondary", class: "bg-secondary" },
  { name: "Muted", var: "--muted", class: "bg-muted" },
  { name: "Accent", var: "--accent", class: "bg-accent" },
  { name: "Success", var: "--success", class: "bg-success" },
  { name: "Warning", var: "--warning", class: "bg-warning" },
  { name: "Danger", var: "--danger", class: "bg-danger" },
  { name: "Border", var: "--border", class: "bg-border" },
  { name: "Surface", var: "--surface", class: "bg-surface" },
];

const TYPE_SCALE = [
  { name: "2xs", class: "text-2xs", px: "11px" },
  { name: "xs", class: "text-xs", px: "12px" },
  { name: "label", class: "text-label", px: "13px" },
  { name: "sm", class: "text-sm", px: "14px" },
  { name: "base", class: "text-base", px: "16px" },
  { name: "lg", class: "text-lg", px: "18px" },
  { name: "xl", class: "text-xl", px: "20px" },
  { name: "2xl", class: "text-2xl", px: "24px" },
];

const SPACING = [
  { name: "1", px: "4px" },
  { name: "2", px: "8px" },
  { name: "3", px: "12px" },
  { name: "4", px: "16px" },
  { name: "6", px: "24px" },
  { name: "8", px: "32px" },
  { name: "10", px: "40px" },
  { name: "12", px: "48px" },
  { name: "16", px: "64px" },
];

export function DesignSystemPage() {
  const [activeTab, setActiveTab] = useState(0);

  return (
    <div className="flex flex-col">
      <TopBar crumbs={["Design System"]} />
      <div className="flex flex-col gap-d-gap overflow-auto p-d-page">

      <h1 className="sr-only">Design System</h1>

      <section>
        <h2 className="mb-4 text-lg font-semibold">Color Palette</h2>
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
          {COLORS.map((c) => (
            <div key={c.name} className="space-y-1.5">
              <div className={`h-16 rounded-lg border border-border ${c.class}`} />
              <div className="text-xs font-medium">{c.name}</div>
              <div className="font-mono text-2xs text-muted-foreground">{c.var}</div>
            </div>
          ))}
        </div>
      </section>

      <section>
        <h2 className="mb-4 text-lg font-semibold">Typography</h2>
        <div className="space-y-3 rounded-xl border border-border p-d-card">
          {TYPE_SCALE.map((t) => (
            <div key={t.name} className="flex items-baseline gap-4">
              <span className="w-16 shrink-0 font-mono text-2xs text-muted-foreground">{t.name} / {t.px}</span>
              <span className={t.class}>The quick brown fox jumps over the lazy dog</span>
            </div>
          ))}
          <div className="mt-4 border-t border-border pt-4">
            <div className="flex items-baseline gap-4">
              <span className="w-16 shrink-0 font-mono text-2xs text-muted-foreground">mono</span>
              <span className="font-mono text-sm">0123456789 ABCDEF const fn = () =&gt; {"{}"}</span>
            </div>
          </div>
        </div>
      </section>

      <section>
        <h2 className="mb-4 text-lg font-semibold">Spacing</h2>
        <div className="space-y-2 rounded-xl border border-border p-d-card">
          {SPACING.map((s) => (
            <div key={s.name} className="flex items-center gap-3">
              <span className="w-12 shrink-0 font-mono text-2xs text-muted-foreground">{s.name} / {s.px}</span>
              <div
                className="h-3 rounded bg-primary/20"
                style={{ width: s.px }}
              />
            </div>
          ))}
        </div>
      </section>

      <section>
        <h2 className="mb-4 text-lg font-semibold">Button</h2>
        <div className="space-y-4">
          <div className="flex flex-wrap gap-3">
            <Button variant="default">Default</Button>
            <Button variant="secondary">Secondary</Button>
            <Button variant="outline">Outline</Button>
            <Button variant="ghost">Ghost</Button>
            <Button variant="destructive">Destructive</Button>
            <Button variant="success">Success</Button>
            <Button variant="link">Link</Button>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <Button size="sm">Small</Button>
            <Button size="default">Default</Button>
            <Button size="lg">Large</Button>
            <Button size="icon"><Search className="h-4 w-4" /></Button>
          </div>
          <div className="flex flex-wrap gap-3">
            <Button loading>Loading</Button>
            <Button disabled>Disabled</Button>
          </div>
        </div>
      </section>

      <section>
        <h2 className="mb-4 text-lg font-semibold">Badges</h2>
        <div className="flex flex-wrap gap-2">
          <Badge tone="neutral">Neutral</Badge>
          <Badge tone="success">Success</Badge>
          <Badge tone="warning">Warning</Badge>
          <Badge tone="danger">Danger</Badge>
          <Badge tone="accent">Accent</Badge>
        </div>
      </section>

      <section>
        <h2 className="mb-4 text-lg font-semibold">Severity</h2>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2"><Severity level="error" /> Error</div>
          <div className="flex items-center gap-2"><Severity level="warning" /> Warning</div>
          <div className="flex items-center gap-2"><Severity level="note" /> Note</div>
        </div>
      </section>

      <section>
        <h2 className="mb-4 text-lg font-semibold">Form Controls</h2>
        <div className="max-w-sm space-y-3">
          <div className="space-y-1.5">
            <Label htmlFor="demo-input">Label</Label>
            <Input id="demo-input" placeholder="Placeholder text..." />
          </div>
          <div className="flex items-center gap-2">
            <Toggle checked={false} onChange={() => {}} />
            <span className="text-sm">Toggle</span>
          </div>
        </div>
      </section>

      <section>
        <h2 className="mb-4 text-lg font-semibold">Tabs</h2>
        <div className="space-y-4">
          <div>
            <div className="mb-1 text-xs text-muted-foreground">Underline variant</div>
            <Tabs items={["Overview", "Details", "Settings"]} active={activeTab} onChange={setActiveTab} />
          </div>
          <div>
            <div className="mb-1 text-xs text-muted-foreground">Pill variant</div>
            <Tabs items={["Overview", "Details", "Settings"]} active={activeTab} onChange={setActiveTab} variant="pill" />
          </div>
        </div>
      </section>

      <section>
        <h2 className="mb-4 text-lg font-semibold">Progress</h2>
        <div className="max-w-md space-y-3">
          <Progress value={85} showLabel />
          <Progress value={65} showLabel />
          <Progress value={40} showLabel />
        </div>
      </section>

      <section>
        <h2 className="mb-4 text-lg font-semibold">Cards</h2>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <Card>
            <CardHeader><CardTitle>Default Card</CardTitle></CardHeader>
            <CardContent><p className="text-sm text-muted-foreground">No shadow, flat appearance</p></CardContent>
          </Card>
          <Card elevated>
            <CardHeader><CardTitle>Elevated Card</CardTitle></CardHeader>
            <CardContent><p className="text-sm text-muted-foreground">With shadow and hover lift</p></CardContent>
          </Card>
        </div>
      </section>

      <section>
        <h2 className="mb-4 text-lg font-semibold">Skeleton Loaders</h2>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
          <SkeletonCard />
          <div className="space-y-2 rounded-xl border border-border p-4">
            <Skeleton className="h-4 w-3/4" />
            <SkeletonText lines={3} />
          </div>
        </div>
      </section>

      <section>
        <h2 className="mb-4 text-lg font-semibold">Empty State</h2>
        <EmptyState
          icon={FolderKanban}
          title="No items yet"
          description="Get started by creating your first item."
          action={{ label: "Create item", onClick: () => {} }}
        />
      </section>

      <section>
        <h2 className="mb-4 text-lg font-semibold">Elevation</h2>
        <div className="flex flex-wrap gap-6">
          <div className="flex h-24 w-32 items-center justify-center rounded-xl border border-border bg-card shadow-raised">
            <span className="text-2xs text-muted-foreground">Raised</span>
          </div>
          <div className="flex h-24 w-32 items-center justify-center rounded-xl border border-border bg-card shadow-floating">
            <span className="text-2xs text-muted-foreground">Floating</span>
          </div>
          <div className="flex h-24 w-32 items-center justify-center rounded-xl border border-border bg-card shadow-overlay">
            <span className="text-2xs text-muted-foreground">Overlay</span>
          </div>
        </div>
      </section>
      </div>
    </div>
  );
}
