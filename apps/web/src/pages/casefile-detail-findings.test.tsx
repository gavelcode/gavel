import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@/test/msw-server";
import { CaseFileDetailPage } from "./casefile-detail";
import { Routes, Route, MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthContext } from "@/entities/user/use-auth";
import { ThemeProvider } from "@/shared/lib/theme";
import { render } from "@testing-library/react";

function renderDetailPage(id = "cf-aaa-111") {
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
          <MemoryRouter initialEntries={[`/casefiles/${id}`]}>
            <Routes>
              <Route path="/casefiles/:id" element={<CaseFileDetailPage />} />
            </Routes>
          </MemoryRouter>
        </ThemeProvider>
      </AuthContext.Provider>
    </QueryClientProvider>,
  );
}

describe("CaseFile Detail — Findings", () => {
  it("should load casefile and findings from API", async () => {
    renderDetailPage();
    expect(await screen.findByText("UnusedVariable")).toBeInTheDocument();
    expect(screen.getByText("NP_NULL_ON_SOME_PATH")).toBeInTheDocument();
    expect(screen.getByText("EmptyCatchBlock")).toBeInTheDocument();
  });

  it("should show commit SHA, branch, verdict in header", async () => {
    renderDetailPage();
    expect(await screen.findByText("abc1234")).toBeInTheDocument();
    expect(screen.getByText("main")).toBeInTheDocument();
    expect(screen.getByText("Passed")).toBeInTheDocument();
  });

  it("should show findings table with severity badges", async () => {
    renderDetailPage();
    await screen.findByText("UnusedVariable");

    const warningBadges = screen.getAllByText("warning");
    expect(warningBadges.length).toBeGreaterThanOrEqual(1);
    const errorBadges = screen.getAllByText("error");
    expect(errorBadges.length).toBeGreaterThanOrEqual(2);
  });

  it("should filter findings by severity", async () => {
    const user = userEvent.setup();
    renderDetailPage();
    await screen.findByText("UnusedVariable");

    const severityDropdown = screen.getByLabelText("Filter by severity");
    await user.selectOptions(severityDropdown, "error");

    expect(screen.queryByText("UnusedVariable")).not.toBeInTheDocument();
    expect(screen.getByText("NP_NULL_ON_SOME_PATH")).toBeInTheDocument();
    expect(screen.getByText("EmptyCatchBlock")).toBeInTheDocument();
  });

  it("should filter findings by tool", async () => {
    const user = userEvent.setup();
    renderDetailPage();
    await screen.findByText("UnusedVariable");

    const toolDropdown = screen.getByLabelText("Filter by tool");
    await user.selectOptions(toolDropdown, "spotbugs");

    expect(screen.queryByText("UnusedVariable")).not.toBeInTheDocument();
    expect(screen.getByText("NP_NULL_ON_SOME_PATH")).toBeInTheDocument();
    expect(screen.queryByText("EmptyCatchBlock")).not.toBeInTheDocument();
  });

  it("should show finding count and stats", async () => {
    renderDetailPage();
    await screen.findByText("UnusedVariable");
    expect(screen.getByText("3 finding(s)")).toBeInTheDocument();
  });

  it("should filter findings by file path input", async () => {
    const user = userEvent.setup();
    renderDetailPage();
    await screen.findByText("UnusedVariable");

    const fileInput = screen.getByLabelText("Filter by file path");
    await user.type(fileInput, "auth");

    expect(screen.queryByText("UnusedVariable")).not.toBeInTheDocument();
    expect(screen.getByText("NP_NULL_ON_SOME_PATH")).toBeInTheDocument();
  });
});
