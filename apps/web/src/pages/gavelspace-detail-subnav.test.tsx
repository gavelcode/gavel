import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Routes, Route } from "react-router-dom";
import { renderApp } from "@/test/render";
import "@/test/msw-server";
import { GavelspaceDetailPage } from "./gavelspace-detail";
import { OverviewTab } from "./gavelspace/overview";
import { ProjectsTab } from "./gavelspace/projects";
import { FindingsTab } from "./gavelspace/findings";
import { PRChecksTab } from "./gavelspace/pr-checks";
import { CaseFilesTab } from "./gavelspace/case-files";

function renderGavelspaceRoutes(route: string) {
  return renderApp(
    <Routes>
      <Route path="/gavelspaces/:name" element={<GavelspaceDetailPage />}>
        <Route index element={<OverviewTab />} />
        <Route path="projects" element={<ProjectsTab />} />
        <Route path="findings" element={<FindingsTab />} />
        <Route path="pr-checks" element={<PRChecksTab />} />
        <Route path="case-files" element={<CaseFilesTab />} />
      </Route>
    </Routes>,
    { route },
  );
}

describe("GavelspaceDetailPage sub-nav shell", () => {
  it("renders 5 tab links in the expected order", async () => {
    renderGavelspaceRoutes("/gavelspaces/alpha");
    await waitFor(() => expect(screen.getByRole("link", { name: /^Overview$/ })).toBeInTheDocument());

    const tabs = screen.getAllByRole("link").filter((el) => {
      const t = el.textContent ?? "";
      return ["Overview", "Projects", "Findings", "PR Checks", "Case Files"].includes(t);
    });
    expect(tabs.map((t) => t.textContent)).toEqual([
      "Overview",
      "Projects",
      "Findings",
      "PR Checks",
      "Case Files",
    ]);
  });

  it("navigates to /projects sub-route on click", async () => {
    const user = userEvent.setup();
    renderGavelspaceRoutes("/gavelspaces/alpha");
    await waitFor(() => expect(screen.getByRole("link", { name: /^Projects$/ })).toBeInTheDocument());

    await user.click(screen.getByRole("link", { name: /^Projects$/ }));

    expect(await screen.findByTestId("gs-tab-projects")).toBeInTheDocument();
  });

  it("renders Overview placeholder by default (no sub-route)", async () => {
    renderGavelspaceRoutes("/gavelspaces/alpha");
    expect(await screen.findByTestId("gs-tab-overview")).toBeInTheDocument();
  });

  it("renders Findings placeholder on /findings", async () => {
    renderGavelspaceRoutes("/gavelspaces/alpha/findings");
    expect(await screen.findByTestId("gs-tab-findings")).toBeInTheDocument();
  });

  it("renders PR Checks placeholder on /pr-checks", async () => {
    renderGavelspaceRoutes("/gavelspaces/alpha/pr-checks");
    expect(await screen.findByTestId("gs-tab-pr-checks")).toBeInTheDocument();
  });

  it("renders Case Files placeholder on /case-files", async () => {
    renderGavelspaceRoutes("/gavelspaces/alpha/case-files");
    expect(await screen.findByTestId("gs-tab-case-files")).toBeInTheDocument();
  });
});
