import { screen } from "@testing-library/react";
import { Routes, Route } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { renderApp } from "@/test/render";
import { server } from "@/test/msw-server";
import { GavelspaceDetailPage } from "../gavelspace-detail";
import { OverviewTab } from "./overview";
import { ProjectStrip } from "./project-strip";
import { Sparkline } from "./sparkline";
import { ActivityFeed } from "./activity-feed";

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

function withGreenCasefiles() {
  const recent = new Date().toISOString();
  server.use(
    http.get("/api/v1/casefiles", () =>
      HttpResponse.json({
        items: [
          { id: "cf-A", project_id: "p1", commit_sha: "aaaaaaa", branch: "main", started_at: recent, verdict_outcome: "pass", total_findings: 2, new_findings: 0, existing_findings: 2, resolved_findings: 0, coverage_percent: 90, created_at: recent },
          { id: "cf-B", project_id: "p2", commit_sha: "bbbbbbb", branch: "main", started_at: recent, verdict_outcome: "pass", total_findings: 1, new_findings: 0, existing_findings: 1, resolved_findings: 0, coverage_percent: 92, created_at: recent },
        ],
        total: 2,
      }),
    ),
  );
}

describe("OverviewTab — green state (Phase 15)", () => {
  it("renders the gavelspace verdict pill as 'Passed' when every project passes", async () => {
    withGreenCasefiles();
    renderOverview();
    expect(await screen.findByTestId("overview-verdict-pill")).toHaveTextContent(/passed/i);
  });

  it("renders a 'last activity' timestamp", async () => {
    withGreenCasefiles();
    renderOverview();
    const lastActivity = await screen.findByTestId("overview-last-activity");
    expect(lastActivity.textContent ?? "").toMatch(/(ago|seconds|minutes|hours|days|just now)/i);
  });

  it("renders the empty state when no case files exist yet", async () => {
    server.use(
      http.get("/api/v1/casefiles", () => HttpResponse.json({ items: [], total: 0 })),
    );
    renderOverview();
    expect(await screen.findByText(/no case files yet/i)).toBeInTheDocument();
  });
});

describe("ProjectStrip", () => {
  const projects = [
    { id: "p1", key: "payment", name: "Payment", latestVerdict: "pass" },
    { id: "p2", key: "auth", name: "Auth", latestVerdict: "pass" },
  ];

  it("renders a card per project with a verdict pill", () => {
    renderApp(<ProjectStrip projects={projects} />);
    expect(screen.getAllByTestId("project-strip-card")).toHaveLength(2);
    expect(screen.getAllByText(/passed/i).length).toBeGreaterThanOrEqual(2);
  });

  it("applies the calm color class when all projects pass", () => {
    renderApp(<ProjectStrip projects={projects} />);
    const cards = screen.getAllByTestId("project-strip-card");
    expect(cards[0].className).toMatch(/calm/);
  });
});

describe("Sparkline", () => {
  it("renders an svg path derived from the 7-day series", () => {
    const series = [
      { day: "2025-03-01", count: 3 },
      { day: "2025-03-02", count: 4 },
      { day: "2025-03-03", count: 2 },
      { day: "2025-03-04", count: 5 },
      { day: "2025-03-05", count: 6 },
      { day: "2025-03-06", count: 3 },
      { day: "2025-03-07", count: 4 },
    ];
    const { container } = renderApp(<Sparkline series={series} />);
    expect(container.querySelector("svg")).not.toBeNull();
    const polyline = container.querySelector("polyline, path");
    expect(polyline).not.toBeNull();
  });
});

describe("ActivityFeed", () => {
  const casefiles = [
    { id: "cf-1", projectId: "p1", commitSha: "aaaaaaa", branch: "main", startedAt: "2025-03-05T10:00:00Z", verdictOutcome: "pass", totalFindings: 1, newFindings: 0, existingFindings: 1, resolvedFindings: 0, coveragePercent: 90, createdAt: "2025-03-05T10:00:00Z" },
    { id: "cf-2", projectId: "p1", commitSha: "bbbbbbb", branch: "main", startedAt: "2025-03-04T10:00:00Z", verdictOutcome: "pass", totalFindings: 2, newFindings: 0, existingFindings: 2, resolvedFindings: 0, coveragePercent: 88, createdAt: "2025-03-04T10:00:00Z" },
  ];

  it("renders one row per case file as a deep link", () => {
    renderApp(<ActivityFeed casefiles={casefiles} />);
    const rows = screen.getAllByTestId("activity-feed-row");
    expect(rows).toHaveLength(2);
    expect(rows[0]).toHaveAttribute("href", "/casefiles/cf-1");
    expect(rows[1]).toHaveAttribute("href", "/casefiles/cf-2");
  });

  it("shows rows in newest-first order (matches input order)", () => {
    renderApp(<ActivityFeed casefiles={casefiles} />);
    const rows = screen.getAllByTestId("activity-feed-row");
    expect(rows[0].textContent ?? "").toContain("aaaaaaa".slice(0, 7));
    expect(rows[1].textContent ?? "").toContain("bbbbbbb".slice(0, 7));
  });
});
