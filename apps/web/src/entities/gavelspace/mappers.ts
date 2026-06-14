import type { components } from "@/shared/api/v1.gen";
import type { GavelspaceSummary, GavelspaceDetail, ProjectRef } from "./model";

export type ProjectRefDTO = components["schemas"]["GavelspaceProject"];
export type GavelspaceSummaryDTO = components["schemas"]["Gavelspace"];
export type GavelspaceDetailDTO = components["schemas"]["GavelspaceDetail"];

function toProjectRef(dto: ProjectRefDTO): ProjectRef {
  return {
    id: dto.id,
    key: dto.key,
    name: dto.name,
    latestVerdict: dto.latest_verdict,
  };
}

export function toGavelspaceSummary(dto: GavelspaceSummaryDTO): GavelspaceSummary {
  return {
    name: dto.name,
    projectCount: dto.project_count,
    createdAt: dto.created_at,
  };
}

export function toGavelspaceDetail(dto: GavelspaceDetailDTO): GavelspaceDetail {
  return {
    name: dto.name,
    projects: dto.projects?.map(toProjectRef) ?? [],
    createdAt: dto.created_at,
  };
}
