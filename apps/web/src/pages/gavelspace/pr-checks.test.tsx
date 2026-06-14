import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Routes, Route } from "react-router-dom";
import { renderApp } from "@/test/render";
import "@/test/msw-server";
import { GavelspaceDetailPage } from "../gavelspace-detail";
import { PRChecksTab } from "./pr-checks";

function renderPRChecksTab(route: string) {
  return renderApp(
    <Routes>
      <Route path="/gavelspaces/:name" element={<GavelspaceDetailPage />}>
        <Route path="pr-checks" element={<PRChecksTab />} />
      </Route>
      <Route path="/pr-checks/:id" element={<div data-testid="pr-detail-stub" />} />
    </Routes>,
    { route },
  );
}

describe("PRChecksTab (Phase 12)", () => {
  it("shows PRs scoped to the current gavelspace", async () => {
    renderPRChecksTab("/gavelspaces/alpha/pr-checks");
    expect(await screen.findByText(/posting pipeline/i)).toBeInTheDocument();
  });

  it("does not leak PRs from other gavelspaces", async () => {
    renderPRChecksTab("/gavelspaces/alpha/pr-checks");
    await screen.findByText(/posting pipeline/i);
    expect(screen.queryByText(/retry logic to payment webhook/i)).not.toBeInTheDocument();
  });

  it("navigates to PR detail when clicking a PR card", async () => {
    const user = userEvent.setup();
    renderPRChecksTab("/gavelspaces/alpha/pr-checks");
    const prTitle = await screen.findByText(/posting pipeline/i);
    await user.click(prTitle.closest("[role='link']")!);
    expect(screen.getByTestId("pr-detail-stub")).toBeInTheDocument();
  });
});
