import { readFileSync } from "node:fs";
import { resolve } from "node:path";

describe("legacy /issues route removal (Phase 11)", () => {
  const routerSrc = readFileSync(
    resolve(__dirname, "..", "app", "router.tsx"),
    "utf-8",
  );

  it("no longer registers the /issues route", () => {
    expect(routerSrc).not.toMatch(/path="\/issues"/);
  });

  it("no longer imports the IssuesPage component", () => {
    expect(routerSrc).not.toMatch(/IssuesPage/);
  });
});
