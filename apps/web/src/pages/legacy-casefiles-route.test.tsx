import { readFileSync } from "node:fs";
import { resolve } from "node:path";

describe("legacy /casefiles list route removal (Phase 13)", () => {
  const routerSrc = readFileSync(
    resolve(__dirname, "..", "app", "router.tsx"),
    "utf-8",
  );

  it("no longer registers the bare /casefiles list route", () => {
    expect(routerSrc).not.toMatch(/path="\/casefiles"\s/);
  });

  it("no longer imports the CaseFilesPage list component", () => {
    expect(routerSrc).not.toMatch(/CaseFilesPage\b/);
  });

  it("keeps the detail route /casefiles/:id", () => {
    expect(routerSrc).toMatch(/path="\/casefiles\/:id"/);
  });

  it("keeps the diff route /casefiles/:id/diff/:compareId", () => {
    expect(routerSrc).toMatch(/path="\/casefiles\/:id\/diff\/:compareId"/);
  });
});
