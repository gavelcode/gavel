export interface CaseFile {
  id: string;
  projectId: string;
  commitSha: string;
  branch: string;
  startedAt: string;
  verdictOutcome: string;
  totalFindings: number;
  newFindings: number;
  existingFindings: number;
  resolvedFindings: number;
  coveragePercent: number | null;
  createdAt: string;
}

export interface EvidenceSummary {
  id: string;
  subtype: string;
  source: string;
  collectedAt: string;
}

export interface RulingView {
  subtype: string;
  passed: boolean;
  detail: string;
}

export interface CaseFileDetail extends CaseFile {
  evidences: EvidenceSummary[];
  rulings: RulingView[];
}
