import type { components } from "@/shared/api/v1.gen";
import type { Pleading, GateResult, GateCondition } from "./model";

export type GateConditionDTO = components["schemas"]["GateCondition"];
export type GateResultDTO = components["schemas"]["GateResult"];
export type PleadingDTO = components["schemas"]["Pleading"];

function toGateCondition(dto: GateConditionDTO): GateCondition {
  return {
    label: dto.label,
    operator: dto.operator,
    value: dto.value,
    threshold: dto.threshold,
    passed: dto.passed,
  };
}

function toGateResult(dto: GateResultDTO): GateResult {
  return {
    passed: dto.passed,
    conditions: dto.conditions?.map(toGateCondition) ?? [],
  };
}

export function toPleading(dto: PleadingDTO): Pleading {
  return {
    id: dto.id,
    projectId: dto.project_id,
    number: dto.number,
    title: dto.title,
    petitioner: dto.petitioner,
    sourceBranch: dto.source_branch,
    targetBranch: dto.target_branch,
    commitSha: dto.commit_sha,
    status: dto.status,
    gateResult: dto.gate_result ? toGateResult(dto.gate_result) : null,
    createdAt: dto.created_at,
    updatedAt: dto.updated_at,
  };
}
