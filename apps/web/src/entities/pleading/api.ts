import { request } from "@/shared/api/client";
import { listResource, type ListOptions } from "@/shared/api/pagination";
import { v1Path, type PaginatedResponse } from "@/shared/api/types";
import type { Pleading } from "./model";
import { type PleadingDTO, toPleading } from "./mappers";

export interface PleadingFilters extends ListOptions {
  projectId?: string;
  status?: string;
  gavelspace?: string;
}

export async function listPleadings(
  filters: PleadingFilters = {},
): Promise<PaginatedResponse<Pleading>> {
  const { projectId, status, gavelspace, ...page } = filters;
  return listResource<PleadingDTO, Pleading>(
    {
      subpath: "/pleadings",
      extraQuery: { project_id: projectId, status, gavelspace },
    },
    toPleading,
    page,
  );
}

export async function getPleading(id: string): Promise<Pleading> {
  const dto = await request<PleadingDTO>(v1Path(`/pleadings/${id}`));
  return toPleading(dto);
}
