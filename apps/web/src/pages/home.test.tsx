import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Routes, Route } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { renderApp } from "@/test/render";
import { server } from "@/test/msw-server";
import { HomePage } from "./home";

function gavelspaceListResponse(names: string[]) {
  return {
    items: names.map((name, i) => ({
      name,
      project_count: i + 1,
      created_at: "2025-01-15T10:00:00Z",
    })),
    total: names.length,
  };
}

function renderHome() {
  return renderApp(
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/gavelspaces/:name" element={<div data-testid="gs-detail-stub" />} />
    </Routes>,
    { route: "/" },
  );
}

describe("HomePage (Phase 14)", () => {
  it("renders a panorama card per gavelspace when N > 1", async () => {
    server.use(
      http.get("/api/v1/gavelspaces", () =>
        HttpResponse.json(gavelspaceListResponse(["alpha", "beta", "gamma"])),
      ),
    );

    renderHome();

    expect(await screen.findByText("alpha")).toBeInTheDocument();
    expect(screen.getByText("beta")).toBeInTheDocument();
    expect(screen.getByText("gamma")).toBeInTheDocument();
  });

  it("redirects to the single gavelspace detail when N = 1", async () => {
    server.use(
      http.get("/api/v1/gavelspaces", () =>
        HttpResponse.json(gavelspaceListResponse(["only-one"])),
      ),
    );

    renderHome();

    await waitFor(() => {
      expect(screen.getByTestId("gs-detail-stub")).toBeInTheDocument();
    });
  });

  it("shows empty state when N = 0", async () => {
    server.use(
      http.get("/api/v1/gavelspaces", () => HttpResponse.json({ items: [], total: 0 })),
    );

    renderHome();

    expect(
      await screen.findByText(/no gavelspaces yet/i),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/gavel judge --server/i),
    ).toBeInTheDocument();
  });

  it("navigates to gavelspace detail when clicking a card", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("/api/v1/gavelspaces", () =>
        HttpResponse.json(gavelspaceListResponse(["alpha", "beta"])),
      ),
    );

    renderHome();

    const alphaCard = await screen.findByText("alpha");
    await user.click(alphaCard.closest("[role='link']")!);

    await waitFor(() => {
      expect(screen.getByTestId("gs-detail-stub")).toBeInTheDocument();
    });
  });
});
