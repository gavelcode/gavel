import type { components } from "@/shared/api/v1.gen";
import type { Finding } from "./model";

export type FindingDTO = components["schemas"]["Finding"];

export function toFinding(dto: FindingDTO): Finding {
  return {
    tool: dto.tool,
    ruleId: dto.rule_id,
    severity: dto.severity as Finding["severity"],
    filePath: dto.file_path,
    line: dto.line,
    message: dto.message,
    fingerprint: dto.fingerprint,
    status: dto.status as Finding["status"],
    source: dto.source,
    commitSha: dto.commit_sha,
    projectKey: dto.project_key,
    casefileId: dto.casefile_id || undefined,
  };
}
