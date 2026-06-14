import { useState, useEffect, useRef, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { search, type SearchResult } from "@/entities/search/api";
import { Input } from "@/shared/ui/input";
import { Spinner } from "@/shared/ui/spinner";
import { FileSearch, FolderOpen, Scale } from "lucide-react";

const TYPE_LABELS: Record<SearchResult["type"], string> = {
  project: "Projects",
  finding: "Findings",
  casefile: "Case Files",
};

const TYPE_ICONS: Record<SearchResult["type"], React.ReactNode> = {
  project: <FolderOpen className="h-3.5 w-3.5" />,
  finding: <FileSearch className="h-3.5 w-3.5" />,
  casefile: <Scale className="h-3.5 w-3.5" />,
};

const TYPE_ORDER: SearchResult["type"][] = ["project", "finding", "casefile"];

interface CommandPaletteProps {
  open: boolean;
  onClose: () => void;
}

export function CommandPalette({ open, onClose }: CommandPaletteProps) {
  if (!open) return null;
  return <CommandPaletteInner onClose={onClose} />;
}

function CommandPaletteInner({ onClose }: { onClose: () => void }) {
  const navigate = useNavigate();
  const inputRef = useRef<HTMLInputElement>(null);
  const dialogRef = useRef<HTMLDivElement>(null);
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(0);

  useEffect(() => {
    setTimeout(() => inputRef.current?.focus(), 0);
  }, []);

  useEffect(() => {
    if (!query.trim()) return;

    const timer = setTimeout(async () => {
      setLoading(true);
      try {
        const data = await search(query);
        setResults(data);
        setSelectedIndex(0);
      } finally {
        setLoading(false);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [query]);

  const displayResults = query.trim() ? results : [];

  const navigateTo = useCallback(
    (url: string) => {
      onClose();
      navigate(url);
    },
    [navigate, onClose],
  );

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Escape") {
      onClose();
    } else if (e.key === "ArrowDown") {
      e.preventDefault();
      setSelectedIndex((i) => Math.min(i + 1, displayResults.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelectedIndex((i) => Math.max(i - 1, 0));
    } else if (e.key === "Enter" && displayResults[selectedIndex]) {
      navigateTo(displayResults[selectedIndex].url);
    } else if (e.key === "Tab") {
      trapFocus(e);
    }
  };

  const trapFocus = (e: React.KeyboardEvent) => {
    const container = dialogRef.current;
    if (!container) return;
    const focusable = container.querySelectorAll<HTMLElement>(
      'input, button, [tabindex]:not([tabindex="-1"])',
    );
    if (focusable.length === 0) return;
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    if (e.shiftKey && document.activeElement === first) {
      e.preventDefault();
      last.focus();
    } else if (!e.shiftKey && document.activeElement === last) {
      e.preventDefault();
      first.focus();
    }
  };

  const grouped = TYPE_ORDER
    .map((type) => ({
      type,
      label: TYPE_LABELS[type],
      icon: TYPE_ICONS[type],
      items: displayResults.filter((r) => r.type === type),
    }))
    .filter((g) => g.items.length > 0);

  let flatIndex = 0;

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center pt-[20vh]"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-label="Search"
      ref={dialogRef}
      onKeyDown={handleKeyDown}
    >
      <div className="fixed inset-0 bg-black/50 animate-fade-in" data-testid="palette-overlay" />
      <div
        className="relative w-full max-w-lg rounded-xl border border-border bg-card shadow-overlay animate-scale-in"
        data-testid="palette-panel"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="border-b border-border p-3">
          <Input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search projects, findings, case files..."
            className="border-0 bg-transparent shadow-none focus-visible:ring-0"
          />
        </div>
        <div className="max-h-[300px] overflow-y-auto p-1.5">
          {loading && <Spinner className="py-6" />}
          {!loading && query.trim() && displayResults.length === 0 && (
            <p className="py-6 text-center text-sm text-muted-foreground">
              No results for &ldquo;{query}&rdquo;
            </p>
          )}
          {!loading &&
            grouped.map((group) => (
              <div key={group.type}>
                <div className="px-2 pb-1 pt-2 text-2xs uppercase tracking-wider text-muted-foreground/60">
                  {group.label}
                </div>
                {group.items.map((item) => {
                  const idx = flatIndex++;
                  return (
                    <button
                      key={`${item.type}-${item.id}`}
                      className={`flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-left text-sm transition-colors duration-fast ${
                        idx === selectedIndex
                          ? "bg-primary/10 text-foreground"
                          : "text-muted-foreground hover:bg-muted/50 hover:text-foreground"
                      }`}
                      onClick={() => navigateTo(item.url)}
                      data-testid="search-result"
                    >
                      <span className="shrink-0 text-muted-foreground/60">{group.icon}</span>
                      <span className="flex-1 truncate">
                        <span className="font-medium text-foreground">{item.title}</span>
                        <span className="ml-2 text-xs text-muted-foreground">{item.subtitle}</span>
                      </span>
                    </button>
                  );
                })}
              </div>
            ))}
        </div>
      </div>
    </div>
  );
}
