import type { components } from "@/shared/api/v1.gen";
import type { ProjectSummary, ProjectDetail, QualityGateRuleView } from "./model";

export type ProjectSummaryDTO = components["schemas"]["Project"];
export type QualityGateRuleViewDTO = components["schemas"]["ProjectRuleView"];
export type ProjectDetailDTO = components["schemas"]["ProjectDetail"];

function toQualityGateRuleView(dto: QualityGateRuleViewDTO): QualityGateRuleView {
  return {
    subtype: dto.subtype,
    strategyType: dto.strategy_type,
  };
}

export function toProjectSummary(dto: ProjectSummaryDTO): ProjectSummary {
  return {
    id: dto.id,
    key: dto.key,
    name: dto.name,
    defaultBranch: dto.default_branch,
    latestVerdict: dto.latest_verdict,
    totalFindings: dto.total_findings,
    createdAt: dto.created_at,
  };
}

export function toProjectDetail(dto: ProjectDetailDTO): ProjectDetail {
  return {
    ...toProjectSummary(dto),
    targetPattern: dto.target_pattern,
    languages: dto.languages,
    qualityGateRules: (dto.quality_gate_rules ?? []).map(toQualityGateRuleView),
    severityCounts: dto.severity_counts ?? {},
  };
}
