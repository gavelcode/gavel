export interface GateCondition {
  label: string;
  operator: string;
  value: string;
  threshold: string;
  passed: boolean;
}

export interface GateResult {
  passed: boolean;
  conditions: GateCondition[];
}

export interface Pleading {
  id: string;
  projectId: string;
  number: number;
  title: string;
  petitioner: string;
  sourceBranch: string;
  targetBranch: string;
  commitSha: string;
  status: string;
  gateResult: GateResult | null;
  createdAt: string;
  updatedAt: string;
}
