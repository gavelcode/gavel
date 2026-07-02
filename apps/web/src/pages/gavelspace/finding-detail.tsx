import { useEffect, useMemo, useRef } from "react";
import { Link, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import * as findingApi from "@/entities/finding/api";
import {
  fetchSource,
  fetchSourceWithContext,
  SourceNotFoundError,
} from "@/entities/source/api";
import { CodeViewer } from "@/features/code-viewer/code-viewer";
import { inferLanguageFromExtension } from "@/features/code-viewer/syntax";
import { Badge } from "@/shared/ui/badge";
import { Severity } from "@/shared/ui/severity";
import { Spinner } from "@/shared/ui/spinner";
import { EmptyState } from "@/shared/ui/empty-state";

export function FindingDetailPage() {
  const { name, findingId } = useParams<{ name: string; findingId: string }>();
  const viewerRef = useRef<HTMLDivElement>(null);

  const findingsQuery = useQuery({
    queryKey: ["gavelspace-findings", name],
    queryFn: () => findingApi.listGlobalFindings({ gavelspace: name }),
    enabled: !!name,
  });

  const finding = findingsQuery.data?.items.find(
    (f) => f.fingerprint === findingId,
  );

  const hasCasefile = !!finding?.casefileId;

  const sourceQuery = useQuery({
    queryKey: [
      "source",
      finding?.projectKey,
      finding?.commitSha,
      finding?.filePath,
      finding?.casefileId,
    ],
    queryFn: () =>
      hasCasefile
        ? fetchSourceWithContext(
            finding.projectKey,
            finding.commitSha,
            finding.filePath,
            finding.casefileId!,
          ).then((r) => ({ content: r.content, coverage: r.coverage }))
        : fetchSource(
            finding!.projectKey,
            finding!.commitSha,
            finding!.filePath,
          ).then((content) => ({ content, coverage: undefined })),
    enabled: !!finding,
    retry: false,
  });

  const sourceContent = sourceQuery.data?.content;
  const coverageMap = useMemo(
    () => sourceQuery.data?.coverage,
    [sourceQuery.data],
  );

  useEffect(() => {
    if (!sourceContent || !viewerRef.current) return;
    const row = viewerRef.current.querySelector(
      "[data-active='true']",
    );
    if (row && typeof row.scrollIntoView === "function") {
      row.scrollIntoView({ block: "center" });
    }
  }, [sourceContent]);

  if (findingsQuery.isLoading) {
    return (
      <div className="p-6">
        <Spinner />
      </div>
    );
  }

  if (!finding) {
    return (
      <div className="flex flex-col gap-4 p-6">
        <BackLink name={name} />
        <EmptyState title="Finding not found" description="No finding matches this URL in the current gavelspace." />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 p-6">
      <BackLink name={name} />

      <div>
        <div className="mb-2 flex flex-wrap items-center gap-2">
          <Severity level={finding.severity} />
          <Badge tone="neutral">{finding.ruleId}</Badge>
          <Badge tone="neutral">{finding.tool}</Badge>
          <Badge tone={finding.status === "new" ? "accent" : "neutral"}>
            {finding.status}
          </Badge>
        </div>
        <h2 className="text-xl font-semibold tracking-tight">{finding.message}</h2>
        <div className="mt-1.5 font-mono text-xs text-muted-foreground">
          {finding.filePath} · line {finding.line}
        </div>
      </div>

      {sourceQuery.isLoading && <Spinner />}

      {sourceQuery.isError &&
        (sourceQuery.error instanceof SourceNotFoundError ? (
          <EmptyState
            title="Source not available for this commit"
            description="This case file was submitted before source bundling was enabled, or the file was not included in the bundle."
          />
        ) : (
          <EmptyState
            title="Failed to load source"
            description={sourceQuery.error instanceof Error ? sourceQuery.error.message : "Unknown error"}
          />
        ))}

      {sourceContent !== undefined && (
        <div ref={viewerRef}>
          <CodeViewer
            source={sourceContent}
            findings={[finding]}
            activeLine={finding.line}
            language={inferLanguageFromExtension(finding.filePath)}
            coverage={coverageMap}
          />
        </div>
      )}
    </div>
  );
}

function BackLink({ name }: { name: string | undefined }) {
  return (
    <Link
      to={`/gavelspaces/${name ?? ""}/findings`}
      className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
    >
      <ArrowLeft className="h-3.5 w-3.5" aria-hidden="true" />
      Back to findings
    </Link>
  );
}
