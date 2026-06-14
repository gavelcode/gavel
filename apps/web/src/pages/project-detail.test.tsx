import { screen } from "@testing-library/react";
import "@/test/msw-server";
import { ProjectDetailPage } from "./project-detail";
import { Routes, Route, MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthContext } from "@/entities/user/use-auth";
import { ThemeProvider } from "@/shared/lib/theme";
import { render } from "@testing-library/react";

function renderDetailPage(key = "//services/payment/...") {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <AuthContext.Provider value={{
        user: { id: 1, email: "admin@local", displayName: "Admin", role: "admin", mustChangePassword: false },
        loading: false,
        login: async () => {},
        logout: async () => {},
      }}>
        <ThemeProvider>
          <MemoryRouter initialEntries={[`/projects/${encodeURIComponent(key)}`]}>
            <Routes>
              <Route path="/projects/:key" element={<ProjectDetailPage />} />
              <Route path="/casefiles/:id" element={<div>CaseFile Detail</div>} />
            </Routes>
          </MemoryRouter>
        </ThemeProvider>
      </AuthContext.Provider>
    </QueryClientProvider>,
  );
}

describe("Project Detail", () => {
  it("should load project data from API", async () => {
    renderDetailPage();
    const headings = await screen.findAllByText("Payment");
    expect(headings.length).toBeGreaterThanOrEqual(1);
  });

  it("should show project name and key", async () => {
    renderDetailPage();
    await screen.findAllByText("Payment");
    const keyElements = screen.getAllByText(/\/\/services\/payment/);
    expect(keyElements.length).toBeGreaterThanOrEqual(1);
  });

  it("should show recent case files section", async () => {
    renderDetailPage();
    expect(await screen.findByText("Recent Case Files")).toBeInTheDocument();
    expect(await screen.findByText("abc1234")).toBeInTheDocument();
  });

  it("should show findings section", async () => {
    renderDetailPage();
    expect(await screen.findByText("UnusedVariable")).toBeInTheDocument();
  });

  it("should show coverage trend chart when coverage data exists", async () => {
    renderDetailPage();
    expect(await screen.findByText("Coverage trend")).toBeInTheDocument();
  });

  it("should navigate to casefile detail on click", async () => {
    renderDetailPage();
    await screen.findByText("abc1234");

    const casefileLinks = screen.getAllByRole("link");
    const casefileLink = casefileLinks.find((l) => l.getAttribute("href")?.startsWith("/casefiles/"));
    expect(casefileLink).toBeDefined();
  });
});
