export interface PaginatedResponse<T> {
  items: T[];
  nextCursor: string | null;
}

export function emptyPage<T>(): PaginatedResponse<T> {
  return { items: [], nextCursor: null };
}

const V1_PREFIX = "/api/v1";

export function v1Path(path: string): string {
  if (path.startsWith("/")) {
    return V1_PREFIX + path;
  }
  return `${V1_PREFIX}/${path}`;
}
