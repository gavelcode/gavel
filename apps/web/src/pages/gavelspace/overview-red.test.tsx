import { screen } from "@testing-library/react";
import { Routes, Route } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { renderApp } from "@/test/render";
import { server } from "@/test/msw-server";
import { GavelspaceDetailPage } from "../gavelspace-detail";
import { OverviewTab } from "./overview";
import { RedBanner } from "./red-banner";
import { ExpandedProjectCard } from "./expanded-project-card";
import { BeforeAfterDiff } from "./before-after-diff";

function renderOverview(route = "/gavelspaces/acme-corp") {
  return renderApp(
    <Routes>
      <Route path="/gavelspaces/:name" element={<GavelspaceDetailPage />}>
        <Route index element={<OverviewTab />} />
      </Route>
    </Routes>,
    { route },
  );
}

function gavelspaceWithProjects(projects: { id: string; name: string; latestVerdict: string }[]) {
  return {
    name: "acme-corp",
    projects: projects.map((p) => ({ id: p.id, key: p.name, name: p.name, latest_verdict: p.latestVerdict })),
    created_at: "2025-01-15T10:00:00Z",
  };
}

interface CaseFileMock {
  id: string;
  project_id: string;
  commit_sha: string;
  branch: string;
  verdict_outcome: string;
  total_findings: number;
  new_findings: number;
  created_at: string;
}

function casefilesForProjects(items: CaseFileMock[]) {
  return server.use(
    http.get("/api/v1/casefiles", () =>
      HttpResponse.json({
        items: items.map((it) => ({
          ...it,
          started_at: it.created_at,
          existing_findings: 0,
          resolved_findings: 0,
          coverage_percent: 80,
        })),
        total: items.length,
      }),
    ),
  );
}

describe("OverviewTab — red state (Phase 16)", () => {
  it("shows the red banner when one project is failing", async () => {
    server.use(
      http.get("/api/v1/gavelspaces/:name", () =>
        HttpResponse.json(gavelspaceWithProjects([
          { id: "p1", name: "api-gateway", latestVerdict: "fail" },
          { id: "p2", name: "billing", latestVerdict: "pass" },
        ])),
      ),
    );
    casefilesForProjects([
      { id: "cf-1", project_id: "p1", commit_sha: "abc1234", branch: "main", verdict_outcome: "fail", total_findings: 3, new_findings: 3, created_at: "2025-03-05T10:00:00Z" },
      { id: "cf-prev", project_id: "p1", commit_sha: "old1111", branch: "main", verdict_outcome: "pass", total_findings: 0, new_findings: 0, created_at: "2025-03-04T10:00:00Z" },
      { id: "cf-2", project_id: "p2", commit_sha: "def5678", branch: "main", verdict_outcome: "pass", total_findings: 0, new_findings: 0, created_at: "2025-03-05T09:00:00Z" },
    ]);

    renderOverview();

    const banner = await screen.findByTestId("overview-red-banner");
    expect(banner.textContent ?? "").toMatch(/1 project failing/i);
    expect(banner.textContent ?? "").toContain("api-gateway");
  });

  it("shows the count + comma-separated names when multiple projects fail", async () => {
    server.use(
      http.get("/api/v1/gavelspaces/:name", () =>
        HttpResponse.json(gavelspaceWithProjects([
          { id: "p1", name: "api-gateway", latestVerdict: "fail" },
          { id: "p2", name: "billing", latestVerdict: "fail" },
          { id: "p3", name: "search", latestVerdict: "pass" },
        ])),
      ),
    );
    casefilesForProjects([
      { id: "cf-1", project_id: "p1", commit_sha: "abc1234", branch: "main", verdict_outcome: "fail", total_findings: 3, new_findings: 3, created_at: "2025-03-05T10:00:00Z" },
      { id: "cf-2", project_id: "p2", commit_sha: "ddd1111", branch: "main", verdict_outcome: "fail", total_findings: 2, new_findings: 2, created_at: "2025-03-05T09:30:00Z" },
      { id: "cf-3", project_id: "p3", commit_sha: "eee2222", branch: "main", verdict_outcome: "pass", total_findings: 0, new_findings: 0, created_at: "2025-03-05T09:00:00Z" },
    ]);

    renderOverview();

    const banner = await screen.findByTestId("overview-red-banner");
    expect(banner.textContent ?? "").toMatch(/2 projects failing/i);
    expect(banner.textContent ?? "").toContain("api-gateway");
    expect(banner.textContent ?? "").toContain("billing");
  });

  it("does not render the banner when every project passes", async () => {
    server.use(
      http.get("/api/v1/gavelspaces/:name", () =>
        HttpResponse.json(gavelspaceWithProjects([
          { id: "p1", name: "billing", latestVerdict: "pass" },
        ])),
      ),
    );
    casefilesForProjects([
      { id: "cf-x", project_id: "p1", commit_sha: "aaa", branch: "main", verdict_outcome: "pass", total_findings: 0, new_findings: 0, created_at: "2025-03-05T10:00:00Z" },
    ]);

    renderOverview();
    await screen.findByTestId("gs-tab-overview");
    expect(screen.queryByTestId("overview-red-banner")).not.toBeInTheDocument();
  });

  it("mixed state: only the failing project loses the calm class; passing ones stay calm", async () => {
    server.use(
      http.get("/api/v1/gavelspaces/:name", () =>
        HttpResponse.json(gavelspaceWithProjects([
          { id: "p1", name: "api-gateway", latestVerdict: "fail" },
          { id: "p2", name: "billing", latestVerdict: "pass" },
        ])),
      ),
    );
    casefilesForProjects([
      { id: "cf-1", project_id: "p1", commit_sha: "abc1234", branch: "main", verdict_outcome: "fail", total_findings: 3, new_findings: 3, created_at: "2025-03-05T10:00:00Z" },
      { id: "cf-2", project_id: "p2", commit_sha: "def5678", branch: "main", verdict_outcome: "pass", total_findings: 0, new_findings: 0, created_at: "2025-03-05T09:00:00Z" },
    ]);

    renderOverview();
    const cards = await screen.findAllByTestId("project-strip-card");
    expect(cards).toHaveLength(2);
    const apiCard = cards.find((c) => c.textContent?.includes("api-gateway"))!;
    const billingCard = cards.find((c) => c.textContent?.includes("billing"))!;
    expect(apiCard.className).not.toMatch(/\bcalm\b/);
    expect(billingCard.className).toMatch(/\bcalm\b/);
  });
});

describe("RedBanner", () => {
  it("singular form for one failing project", () => {
    renderApp(<RedBanner names={["api-gateway"]} />);
    const el = screen.getByTestId("overview-red-banner");
    expect(el.textContent ?? "").toMatch(/1 project failing/i);
    expect(el.textContent ?? "").toContain("api-gateway");
  });

  it("plural form with comma-separated names for multiple", () => {
    renderApp(<RedBanner names={["api-gateway", "billing"]} />);
    const el = screen.getByTestId("overview-red-banner");
    expect(el.textContent ?? "").toMatch(/2 projects failing/i);
    expect(el.textContent ?? "").toContain("api-gateway");
    expect(el.textContent ?? "").toContain("billing");
  });
});

describe("ExpandedProjectCard", () => {
  const latest = {
    id: "cf-1", projectId: "p1", commitSha: "abc1234567", branch: "fix/bug",
    startedAt: "2025-03-05T10:00:00Z", verdictOutcome: "fail",
    totalFindings: 5, newFindings: 3, existingFindings: 2, resolvedFindings: 0,
    coveragePercent: 70, createdAt: "2025-03-05T10:00:00Z",
  };

  it("shows project name, short commit SHA, branch, and new-findings count", () => {
    renderApp(<ExpandedProjectCard projectName="api-gateway" latest={latest} previous={null} />);
    expect(screen.getByText("api-gateway")).toBeInTheDocument();
    expect(screen.getByText(/abc1234/)).toBeInTheDocument();
    expect(screen.getByText("fix/bug")).toBeInTheDocument();
    expect(screen.getByText(/3 new findings/i)).toBeInTheDocument();
  });
});

describe("BeforeAfterDiff", () => {
  it("renders 'before' and 'after' findings counts", () => {
    renderApp(<BeforeAfterDiff before={0} after={3} />);
    const el = screen.getByTestId("overview-before-after-diff");
    expect(el.textContent ?? "").toMatch(/before:\s*0/i);
    expect(el.textContent ?? "").toMatch(/after:\s*3/i);
  });
});
