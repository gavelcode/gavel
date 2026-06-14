export interface ProjectRef {
  id: string;
  key: string;
  name: string;
  latestVerdict: string;
}

export interface GavelspaceSummary {
  name: string;
  projectCount: number;
  createdAt: string;
}

export interface GavelspaceDetail {
  name: string;
  projects: ProjectRef[];
  createdAt: string;
}
