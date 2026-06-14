import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Routes, Route } from "react-router-dom";
import { renderApp } from "@/test/render";
import "@/test/msw-server";
import { GavelspaceDetailPage } from "../gavelspace-detail";
import { FindingsTab } from "./findings";

function renderFindingsTab(route: string) {
  return renderApp(
    <Routes>
      <Route path="/gavelspaces/:name" element={<GavelspaceDetailPage />}>
        <Route path="findings" element={<FindingsTab />} />
        <Route path="findings/:fingerprint" element={<FindingsTab />} />
      </Route>
    </Routes>,
    { route },
  );
}

describe("FindingsTab (Phase 11)", () => {
  it("shows findings scoped to the current gavelspace", async () => {
    renderFindingsTab("/gavelspaces/alpha/findings");

    expect((await screen.findAllByText("Variable 'x' is never used")).length).toBeGreaterThan(0);
    expect(screen.getAllByText("Possible null pointer dereference").length).toBeGreaterThan(0);
  });

  it("does not leak findings from other gavelspaces", async () => {
    renderFindingsTab("/gavelspaces/alpha/findings");

    await screen.findAllByText("Variable 'x' is never used");
    expect(screen.queryByText("Empty catch block")).not.toBeInTheDocument();
    expect(screen.queryByText(/Prefer Splitter/)).not.toBeInTheDocument();
  });

  it("filters by severity when clicking a severity badge", async () => {
    const user = userEvent.setup();
    renderFindingsTab("/gavelspaces/alpha/findings");

    await screen.findAllByText("Variable 'x' is never used");

    const errorBadge = screen.getByRole("button", { name: "error" });
    await user.click(errorBadge);

    const count = await screen.findByTestId("finding-count");
    expect(count).toHaveTextContent("1 findings");
    expect(screen.queryByText("Variable 'x' is never used")).not.toBeInTheDocument();
  });

  it("filters by file path when typing in the input", async () => {
    const user = userEvent.setup();
    renderFindingsTab("/gavelspaces/alpha/findings");

    await screen.findAllByText("Variable 'x' is never used");

    const input = screen.getByLabelText("Filter by file path");
    await user.type(input, "auth");

    const count = await screen.findByTestId("finding-count");
    expect(count).toHaveTextContent("1 findings");
    expect(screen.queryByText("Variable 'x' is never used")).not.toBeInTheDocument();
  });

  it("clears filters when clicking Reset", async () => {
    const user = userEvent.setup();
    renderFindingsTab("/gavelspaces/alpha/findings");

    await screen.findAllByText("Variable 'x' is never used");

    const errorBadge = screen.getByRole("button", { name: "error" });
    await user.click(errorBadge);

    await screen.findByTestId("finding-count");
    expect(screen.queryByText("Variable 'x' is never used")).not.toBeInTheDocument();

    const resetButton = screen.getByRole("button", { name: "Reset" });
    await user.click(resetButton);

    expect((await screen.findAllByText("Variable 'x' is never used")).length).toBeGreaterThan(0);
  });

  it("shows detail pane for the default-selected finding", async () => {
    renderFindingsTab("/gavelspaces/alpha/findings");

    await screen.findAllByText("Variable 'x' is never used");

    expect(screen.getByText("Details")).toBeInTheDocument();
    expect(screen.getAllByText("UnusedVariable").length).toBeGreaterThanOrEqual(1);
  });

  it("updates detail pane when clicking a different IssueRow", async () => {
    const user = userEvent.setup();
    renderFindingsTab("/gavelspaces/alpha/findings");

    await screen.findAllByText("Variable 'x' is never used");

    expect(screen.getAllByText("UnusedVariable").length).toBeGreaterThanOrEqual(1);

    const nullPtrLinks = screen.getAllByRole("link").filter(
      (el) => el.textContent?.includes("Possible null pointer dereference"),
    );
    await user.click(nullPtrLinks[0]);

    expect(screen.getAllByText("NP_NULL_ON_SOME_PATH").length).toBeGreaterThanOrEqual(1);
  });
});
