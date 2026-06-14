export interface ProjectSummary {
  id: string;
  key: string;
  name: string;
  defaultBranch: string;
  latestVerdict: string;
  totalFindings: number;
  createdAt: string;
}

export interface QualityGateRuleView {
  subtype: string;
  strategyType: string;
}

export interface ProjectDetail extends ProjectSummary {
  targetPattern: string;
  languages: string[];
  qualityGateRules: QualityGateRuleView[];
  severityCounts: Record<string, number>;
}
