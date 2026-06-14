import { request } from "@/shared/api/client";
import { listResource, type ListOptions } from "@/shared/api/pagination";
import { v1Path, type PaginatedResponse } from "@/shared/api/types";
import type { ProjectSummary, ProjectDetail } from "./model";
import {
  type ProjectSummaryDTO,
  type ProjectDetailDTO,
  toProjectSummary,
  toProjectDetail,
} from "./mappers";

export async function listProjects(
  options: ListOptions = {},
): Promise<PaginatedResponse<ProjectSummary>> {
  return listResource<ProjectSummaryDTO, ProjectSummary>(
    { subpath: "/projects" },
    toProjectSummary,
    options,
  );
}

export async function getProject(key: string): Promise<ProjectDetail> {
  const dto = await request<ProjectDetailDTO>(
    v1Path(`/projects/${encodeURIComponent(key)}`),
  );
  return toProjectDetail(dto);
}
