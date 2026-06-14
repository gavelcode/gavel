import type { components } from "@/shared/api/v1.gen";
import type { CaseFile, CaseFileDetail, EvidenceSummary, RulingView } from "./model";

export type CaseFileDTO = components["schemas"]["CaseFile"];
export type CaseFileDetailDTO = components["schemas"]["CaseFileDetail"];
export type EvidenceSummaryDTO = components["schemas"]["Evidence"];
export type RulingViewDTO = components["schemas"]["Ruling"];

export function toCaseFile(dto: CaseFileDTO): CaseFile {
  return {
    id: dto.id,
    projectId: dto.project_id,
    commitSha: dto.commit_sha,
    branch: dto.branch,
    startedAt: dto.started_at,
    verdictOutcome: dto.verdict_outcome,
    totalFindings: dto.total_findings,
    newFindings: dto.new_findings,
    existingFindings: dto.existing_findings,
    resolvedFindings: dto.resolved_findings,
    coveragePercent: dto.coverage_percent ?? null,
    createdAt: dto.created_at,
  };
}

function toEvidenceSummary(dto: EvidenceSummaryDTO): EvidenceSummary {
  return {
    id: dto.id,
    subtype: dto.subtype,
    source: dto.source,
    collectedAt: dto.collected_at,
  };
}

function toRulingView(dto: RulingViewDTO): RulingView {
  return {
    subtype: dto.subtype,
    passed: dto.passed,
    detail: dto.detail,
  };
}

export function toCaseFileDetail(dto: CaseFileDetailDTO): CaseFileDetail {
  return {
    ...toCaseFile(dto),
    evidences: dto.evidences?.map(toEvidenceSummary) ?? [],
    rulings: dto.rulings?.map(toRulingView) ?? [],
  };
}
