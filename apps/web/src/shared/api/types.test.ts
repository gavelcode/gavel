import { emptyPage, v1Path } from "./types";

describe("emptyPage", () => {
  it("returns an empty paginated response", () => {
    const page = emptyPage();

    expect(page).toEqual({ items: [], nextCursor: null });
  });
});

describe("v1Path", () => {
  it("prepends /api/v1 to a path with leading slash", () => {
    expect(v1Path("/health")).toBe("/api/v1/health");
  });

  it("prepends /api/v1/ to a path without leading slash", () => {
    expect(v1Path("health")).toBe("/api/v1/health");
  });

  it("handles nested paths", () => {
    expect(v1Path("/projects/core")).toBe("/api/v1/projects/core");
  });
});
