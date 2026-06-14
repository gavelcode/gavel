import { request } from "@/shared/api/client";
import { listResource, type ListOptions } from "@/shared/api/pagination";
import { v1Path, type PaginatedResponse } from "@/shared/api/types";
import type { CaseFile } from "./model";
import { type CaseFileDTO, toCaseFile } from "./mappers";

export interface CaseFileFilters extends ListOptions {
  projectId?: string;
  gavelspace?: string;
}

export async function listCaseFiles(
  filters: CaseFileFilters = {},
): Promise<PaginatedResponse<CaseFile>> {
  const { projectId, gavelspace, ...page } = filters;
  return listResource<CaseFileDTO, CaseFile>(
    {
      subpath: "/casefiles",
      extraQuery: { project_id: projectId, gavelspace },
    },
    toCaseFile,
    page,
  );
}

export async function getCaseFile(id: string): Promise<CaseFile> {
  const dto = await request<CaseFileDTO>(v1Path(`/casefiles/${id}`));
  return toCaseFile(dto);
}
