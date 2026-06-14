import { screen } from "@testing-library/react";
import { Routes, Route } from "react-router-dom";
import { renderApp } from "@/test/render";
import "@/test/msw-server";
import { GavelspaceDetailPage } from "../gavelspace-detail";
import { CaseFilesTab } from "./case-files";

function renderCaseFilesTab(route: string) {
  return renderApp(
    <Routes>
      <Route path="/gavelspaces/:name" element={<GavelspaceDetailPage />}>
        <Route path="case-files" element={<CaseFilesTab />} />
      </Route>
    </Routes>,
    { route },
  );
}

describe("CaseFilesTab (Phase 13)", () => {
  it("shows case files scoped to the current gavelspace", async () => {
    renderCaseFilesTab("/gavelspaces/alpha/case-files");
    expect(await screen.findByText(/abc1234/)).toBeInTheDocument();
    expect(screen.getByText(/def5678/)).toBeInTheDocument();
  });

  it("does not leak case files from other gavelspaces", async () => {
    renderCaseFilesTab("/gavelspaces/alpha/case-files");
    await screen.findByText(/abc1234/);
    expect(screen.queryByText(/beta-sha/)).not.toBeInTheDocument();
  });
});
