import { request } from "@/shared/api/client";
import { listResource, type ListOptions } from "@/shared/api/pagination";
import { v1Path, type PaginatedResponse } from "@/shared/api/types";
import type { GavelspaceSummary, GavelspaceDetail } from "./model";
import {
  type GavelspaceSummaryDTO,
  type GavelspaceDetailDTO,
  toGavelspaceSummary,
  toGavelspaceDetail,
} from "./mappers";

export async function listGavelspaces(
  options: ListOptions = {},
): Promise<PaginatedResponse<GavelspaceSummary>> {
  return listResource<GavelspaceSummaryDTO, GavelspaceSummary>(
    { subpath: "/gavelspaces" },
    toGavelspaceSummary,
    options,
  );
}

export async function getGavelspace(name: string): Promise<GavelspaceDetail> {
  const dto = await request<GavelspaceDetailDTO>(
    v1Path(`/gavelspaces/${encodeURIComponent(name)}`),
  );
  return toGavelspaceDetail(dto);
}
