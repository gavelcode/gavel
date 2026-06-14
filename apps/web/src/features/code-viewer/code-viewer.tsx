import { useMemo } from "react";
import type { Finding } from "@/entities/finding/model";
import { CoverageGutter, type CoverageState } from "./coverage-gutter";
import { FindingMarkers } from "./finding-markers";
import { useAsyncSourceHighlights } from "./use-async-source-highlights";

export interface CodeViewerProps {
  source: string;
  activeLine?: number;
  coverage?: Map<number, CoverageState>;
  findings?: Finding[];
  onFindingClick?: (finding: Finding) => void;
  language?: string;
}

function groupFindingsByLine(findings: Finding[]): Map<number, Finding[]> {
  const grouped = new Map<number, Finding[]>();
  for (const finding of findings) {
    const existing = grouped.get(finding.line);
    if (existing) {
      existing.push(finding);
    } else {
      grouped.set(finding.line, [finding]);
    }
  }
  return grouped;
}

export function CodeViewer({
  source,
  activeLine,
  coverage,
  findings,
  onFindingClick,
  language,
}: CodeViewerProps) {
  const findingsByLine = useMemo(
    () => groupFindingsByLine(findings ?? []),
    [findings],
  );
  const highlightedLines = useAsyncSourceHighlights(source, language);

  if (source.length === 0) {
    return (
      <div
        data-testid="code-viewer"
        className="rounded border border-border bg-surface p-6 text-center text-sm text-muted-foreground"
      >
        <span data-testid="code-viewer-empty">Empty file</span>
      </div>
    );
  }

  const lines = source.split("\n");

  return (
    <div
      data-testid="code-viewer"
      className="code-viewer overflow-auto rounded border border-border bg-surface font-mono text-sm leading-5"
    >
      <ol className="m-0 list-none p-0">
        {lines.map((text, idx) => {
          const lineNumber = idx + 1;
          const isActive = activeLine === lineNumber;
          const coverageState: CoverageState =
            coverage?.get(lineNumber) ?? "none";
          const lineFindings = findingsByLine.get(lineNumber) ?? [];
          const highlighted = highlightedLines?.[idx];
          return (
            <li
              key={lineNumber}
              data-testid="code-line"
              data-line-number={lineNumber}
              data-active={isActive ? "true" : undefined}
              className={
                "flex items-stretch " +
                (isActive ? "bg-warning/10" : "hover:bg-muted/40")
              }
            >
              <span
                aria-hidden="true"
                className="select-none px-3 text-right text-muted-foreground tabular-nums"
                style={{ minWidth: "3.5rem" }}
              >
                {lineNumber}
              </span>
              <CoverageGutter state={coverageState} />
              <FindingMarkers
                findings={lineFindings}
                onFindingClick={onFindingClick}
              />
              {highlighted !== undefined ? (
                <span
                  data-testid="code-line-text"
                  className="flex-1 pr-3"
                  style={{ whiteSpace: "pre" }}
                  dangerouslySetInnerHTML={{ __html: highlighted }}
                />
              ) : (
                <span
                  data-testid="code-line-text"
                  className="flex-1 pr-3"
                  style={{ whiteSpace: "pre" }}
                >
                  {text}
                </span>
              )}
            </li>
          );
        })}
      </ol>
    </div>
  );
}
