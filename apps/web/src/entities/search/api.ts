import { request } from "@/shared/api/client";
import { v1Path } from "@/shared/api/types";

export interface SearchResult {
  type: "project" | "finding" | "casefile";
  id: string;
  title: string;
  subtitle: string;
  url: string;
}

export async function search(q: string, limit = 10): Promise<SearchResult[]> {
  const params = new URLSearchParams({ q, limit: String(limit) });
  const res = await request<{ results: SearchResult[] }>(
    `${v1Path("/search")}?${params}`,
  );
  return res.results;
}
