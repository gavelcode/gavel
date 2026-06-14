import { request } from "./client";
import { v1Path, type PaginatedResponse } from "./types";

interface ListWire<TDTO> {
  items: TDTO[];
  next_cursor: string | null;
}

export interface ListOptions {
  limit?: number;
  cursor?: string;
}

export interface ListPath {
  subpath: string;
  extraQuery?: Record<string, string | number | undefined>;
}

export async function listResource<TDTO, T>(
  path: ListPath,
  transform: (dto: TDTO) => T,
  options: ListOptions = {},
): Promise<PaginatedResponse<T>> {
  const params = new URLSearchParams();
  if (options.limit != null) params.set("limit", String(options.limit));
  if (options.cursor) params.set("cursor", options.cursor);
  for (const [key, value] of Object.entries(path.extraQuery ?? {})) {
    if (value === undefined || value === null || value === "") continue;
    params.set(key, String(value));
  }
  const qs = params.toString();
  const url = qs ? `${v1Path(path.subpath)}?${qs}` : v1Path(path.subpath);
  const wire = await request<ListWire<TDTO>>(url);
  const items = wire.items.map(transform);
  return { items, nextCursor: wire.next_cursor };
}
