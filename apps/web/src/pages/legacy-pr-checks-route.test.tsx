import { readFileSync } from "node:fs";
import { resolve } from "node:path";

describe("legacy /pr-checks list route removal (Phase 12)", () => {
  const routerSrc = readFileSync(
    resolve(__dirname, "..", "app", "router.tsx"),
    "utf-8",
  );

  it("no longer registers the /pr-checks list route", () => {
    expect(routerSrc).not.toMatch(/path="\/pr-checks"/);
  });

  it("no longer imports the PRCheckPage list component", () => {
    expect(routerSrc).not.toMatch(/PRCheckPage\b/);
  });
});
