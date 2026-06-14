import { toFinding, type FindingDTO } from "./mappers";

describe("toFinding", () => {
  it("maps snake_case server fields onto camelCase model fields", () => {
    const dto: FindingDTO = {
      tool: "golangci-lint",
      rule_id: "errcheck",
      severity: "error",
      file_path: "cmd/main.go",
      line: 42,
      message: "error return value not checked",
      fingerprint: "abc123def456",
      status: "new",
      source: "sarif",
      commit_sha: "deadbeefcafe",
      project_key: "backend",
    };

    const finding = toFinding(dto);

    expect(finding.tool).toBe("golangci-lint");
    expect(finding.ruleId).toBe("errcheck");
    expect(finding.filePath).toBe("cmd/main.go");
    expect(finding.line).toBe(42);
    expect(finding.fingerprint).toBe("abc123def456");
    expect(finding.commitSha).toBe("deadbeefcafe");
    expect(finding.projectKey).toBe("backend");
  });
});
