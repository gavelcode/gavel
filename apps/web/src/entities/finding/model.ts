export type Severity = "error" | "warning" | "note";

export type FindingStatus = "new" | "existing" | "resolved";

export interface Finding {
  tool: string;
  ruleId: string;
  severity: Severity;
  filePath: string;
  line: number;
  message: string;
  fingerprint: string;
  status: FindingStatus;
  source: string;
  commitSha: string;
  projectKey: string;
  casefileId?: string;
}

export interface FindingFilters {
  projectId?: string;
  casefileId?: string;
  tool?: string;
  severity?: Severity;
  status?: FindingStatus;
  filePath?: string;
  gavelspace?: string;
}
