import { listResource, type ListOptions } from "@/shared/api/pagination";
import { type PaginatedResponse } from "@/shared/api/types";
import type { Finding, FindingFilters } from "./model";
import { type FindingDTO, toFinding } from "./mappers";

export async function listFindings(
  casefileId: string,
  options: ListOptions = {},
): Promise<PaginatedResponse<Finding>> {
  return listResource<FindingDTO, Finding>(
    {
      subpath: "/findings",
      extraQuery: { casefile_id: casefileId },
    },
    toFinding,
    { limit: 100, ...options },
  );
}

export async function listGlobalFindings(
  filters: FindingFilters = {},
  options: ListOptions = {},
): Promise<PaginatedResponse<Finding>> {
  return listResource<FindingDTO, Finding>(
    {
      subpath: "/findings",
      extraQuery: {
        project_id: filters.projectId,
        tool: filters.tool,
        severity: filters.severity,
        status: filters.status,
        file_path: filters.filePath,
        gavelspace: filters.gavelspace,
      },
    },
    toFinding,
    options,
  );
}
