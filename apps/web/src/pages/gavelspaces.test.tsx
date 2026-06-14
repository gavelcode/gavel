import { screen } from "@testing-library/react";
import { render } from "@testing-library/react";
import { Routes, Route, MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthContext } from "@/entities/user/use-auth";
import { ThemeProvider } from "@/shared/lib/theme";
import userEvent from "@testing-library/user-event";
import { renderApp } from "@/test/render";
import "@/test/msw-server";
import { GavelspacesPage } from "./gavelspaces";
import { GavelspaceDetailPage } from "./gavelspace-detail";
import { OverviewTab } from "./gavelspace/overview";

function renderDetailPage(name = "acme-corp") {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <AuthContext.Provider
        value={{
          user: {
            id: 1,
            email: "admin@local",
            displayName: "Admin",
            role: "admin",
            mustChangePassword: false,
          },
          loading: false,
          login: async () => {},
          logout: async () => {},
        }}
      >
        <ThemeProvider>
          <MemoryRouter initialEntries={[`/gavelspaces/${name}`]}>
            <Routes>
              <Route
                path="/gavelspaces/:name"
                element={<GavelspaceDetailPage />}
              >
                <Route index element={<OverviewTab />} />
              </Route>
              <Route
                path="/projects/:key"
                element={<div>Project Detail</div>}
              />
            </Routes>
          </MemoryRouter>
        </ThemeProvider>
      </AuthContext.Provider>
    </QueryClientProvider>,
  );
}

describe("GavelspacesPage", () => {
  it("should display gavelspaces list", async () => {
    renderApp(<GavelspacesPage />);
    expect(await screen.findByText("acme-corp")).toBeInTheDocument();
    expect(screen.getByText("platform")).toBeInTheDocument();
  });

  it("should show project counts", async () => {
    renderApp(<GavelspacesPage />);
    expect(await screen.findByText("3 projects")).toBeInTheDocument();
    expect(screen.getByText("1 project")).toBeInTheDocument();
  });

  it("should navigate to detail on click", async () => {
    const user = userEvent.setup();
    renderApp(<GavelspacesPage />, { route: "/gavelspaces" });
    const card = await screen.findByText("acme-corp");
    await user.click(card);
  });
});

describe("GavelspaceDetailPage", () => {
  it("should display gavelspace with projects", async () => {
    renderDetailPage("acme-corp");
    expect(await screen.findByText("Payment")).toBeInTheDocument();
    expect(screen.getByText("Auth")).toBeInTheDocument();
  });

  it("should show verdict badges for projects", async () => {
    renderDetailPage("acme-corp");
    expect(await screen.findByText("Passed")).toBeInTheDocument();
    expect(screen.getByText("Failed")).toBeInTheDocument();
  });
});
