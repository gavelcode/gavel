import type { CoverageState } from "@/features/code-viewer/coverage-gutter";
import { v1Path } from "@/shared/api/types";

export class SourceNotFoundError extends Error {
  constructor(message = "source not found") {
    super(message);
    this.name = "SourceNotFoundError";
  }
}

export async function fetchSource(
  projectKey: string,
  commit: string,
  path: string,
): Promise<string> {
  const params = new URLSearchParams({ commit, path });
  const res = await fetch(
    `${v1Path(`/projects/${encodeURIComponent(projectKey)}/source`)}?${params}`,
    { credentials: "include" },
  );
  if (res.status === 404) {
    throw new SourceNotFoundError();
  }
  if (!res.ok) {
    throw new Error(`fetch source failed: ${res.status}`);
  }
  return res.text();
}

export interface SourceWithContext {
  content: string;
  coverage: Map<number, CoverageState> | undefined;
}

export async function fetchSourceWithContext(
  projectKey: string,
  commit: string,
  path: string,
  casefileId: string,
): Promise<SourceWithContext> {
  const params = new URLSearchParams({ commit, path, casefile: casefileId });
  const res = await fetch(
    `${v1Path(`/projects/${encodeURIComponent(projectKey)}/source`)}?${params}`,
    { credentials: "include" },
  );
  if (res.status === 404) {
    throw new SourceNotFoundError();
  }
  if (!res.ok) {
    throw new Error(`fetch source failed: ${res.status}`);
  }
  const data = await res.json();
  return {
    content: data.content,
    coverage: buildCoverageMap(data.coverage),
  };
}

function buildCoverageMap(
  coverage: { covered_lines?: number[]; uncovered_lines?: number[] } | undefined,
): Map<number, CoverageState> | undefined {
  if (!coverage) return undefined;
  const map = new Map<number, CoverageState>();
  for (const line of coverage.covered_lines ?? []) {
    map.set(line, "covered");
  }
  for (const line of coverage.uncovered_lines ?? []) {
    map.set(line, "uncovered");
  }
  return map;
}
