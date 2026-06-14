import { toCaseFile, toCaseFileDetail } from "./mappers";
import type { CaseFileDTO, CaseFileDetailDTO } from "./mappers";

describe("toCaseFile", () => {
  it("maps DTO snake_case to model camelCase", () => {
    const dto: CaseFileDTO = {
      id: "cf-1",
      project_id: "proj-1",
      commit_sha: "abc1234",
      branch: "main",
      started_at: "2025-03-01T09:59:00Z",
      verdict_outcome: "pass",
      total_findings: 3,
      new_findings: 1,
      existing_findings: 2,
      resolved_findings: 0,
      coverage_percent: 82.5,
      created_at: "2025-03-01T10:00:00Z",
    };

    const result = toCaseFile(dto);

    expect(result).toEqual({
      id: "cf-1",
      projectId: "proj-1",
      commitSha: "abc1234",
      branch: "main",
      startedAt: "2025-03-01T09:59:00Z",
      verdictOutcome: "pass",
      totalFindings: 3,
      newFindings: 1,
      existingFindings: 2,
      resolvedFindings: 0,
      coveragePercent: 82.5,
      createdAt: "2025-03-01T10:00:00Z",
    });
  });

  it("maps null coverage_percent", () => {
    const dto = {
      id: "cf-2", project_id: "p", commit_sha: "s", branch: "b",
      started_at: "t", verdict_outcome: "fail", total_findings: 0,
      new_findings: 0, existing_findings: 0, resolved_findings: 0,
      created_at: "t",
    } as CaseFileDTO;

    expect(toCaseFile(dto).coveragePercent).toBeNull();
  });
});

describe("toCaseFileDetail", () => {
  it("maps evidences and rulings", () => {
    const dto = {
      id: "cf-1", project_id: "p", commit_sha: "s", branch: "b",
      started_at: "t", verdict_outcome: "pass", total_findings: 0,
      new_findings: 0, existing_findings: 0, resolved_findings: 0,
      coverage_percent: 90, created_at: "t",
      evidences: [
        { id: "ev-1", subtype: "lint", source: "golangci-lint", collected_at: "2025-01-01T00:00:00Z" },
      ],
      rulings: [
        { subtype: "code_quality", passed: true, detail: "0 findings" },
      ],
    } as CaseFileDetailDTO;

    const result = toCaseFileDetail(dto);

    expect(result.evidences).toEqual([
      { id: "ev-1", subtype: "lint", source: "golangci-lint", collectedAt: "2025-01-01T00:00:00Z" },
    ]);
    expect(result.rulings).toEqual([
      { subtype: "code_quality", passed: true, detail: "0 findings" },
    ]);
  });

  it("defaults to empty arrays when evidences and rulings are undefined", () => {
    const dto = {
      id: "cf-1", project_id: "p", commit_sha: "s", branch: "b",
      started_at: "t", verdict_outcome: "pass", total_findings: 0,
      new_findings: 0, existing_findings: 0, resolved_findings: 0,
      coverage_percent: 90, created_at: "t",
    } as CaseFileDetailDTO;

    const result = toCaseFileDetail(dto);

    expect(result.evidences).toEqual([]);
    expect(result.rulings).toEqual([]);
  });
});
