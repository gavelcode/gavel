import { screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/test/msw-server";
import { CaseFileDiffPage } from "./casefile-diff";
import { Routes, Route, MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthContext } from "@/entities/user/use-auth";
import { ThemeProvider } from "@/shared/lib/theme";
import { render } from "@testing-library/react";

function renderDiffPage(baseId = "cf-aaa-111", compareId = "cf-bbb-222") {
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
          <MemoryRouter initialEntries={[`/casefiles/${baseId}/diff/${compareId}`]}>
            <Routes>
              <Route path="/casefiles/:id/diff/:compareId" element={<CaseFileDiffPage />} />
              <Route path="/casefiles/:id" element={<div>CaseFile Detail</div>} />
            </Routes>
          </MemoryRouter>
        </ThemeProvider>
      </AuthContext.Provider>
    </QueryClientProvider>,
  );
}

describe("CaseFileDiffPage", () => {
  beforeEach(() => {
    server.use(
      http.get("/api/v1/casefiles/:id", ({ params }) => {
        if (params.id === "cf-aaa-111") {
          return HttpResponse.json({
            id: "cf-aaa-111", project_id: "proj-aaa-111", commit_sha: "abc1234",
            branch: "main", started_at: "2025-03-01T09:59:00Z", verdict_outcome: "pass",
            total_findings: 3, new_findings: 1, existing_findings: 2, resolved_findings: 0,
            coverage_percent: 82.5, created_at: "2025-03-01T10:00:00Z",
          });
        }
        return HttpResponse.json({
          id: "cf-bbb-222", project_id: "proj-aaa-111", commit_sha: "def5678",
          branch: "main", started_at: "2025-02-28T09:59:00Z", verdict_outcome: "fail",
          total_findings: 5, new_findings: 3, existing_findings: 1, resolved_findings: 1,
          coverage_percent: 78.3, created_at: "2025-02-28T10:00:00Z",
        });
      }),
      http.get("/api/v1/findings", ({ request }) => {
        const url = new URL(request.url);
        const casefileId = url.searchParams.get("casefile_id");
        if (casefileId === "cf-aaa-111") {
          return HttpResponse.json({
            items: [
              { tool: "pmd", rule_id: "UnusedVariable", severity: "warning", file_path: "handler.go", line: 42, message: "Variable x unused", fingerprint: "fp1", status: "existing", source: "lint" },
              { tool: "pmd", rule_id: "EmptyCatch", severity: "error", file_path: "handler.go", line: 87, message: "Empty catch block", fingerprint: "fp2", status: "existing", source: "lint" },
            ],
            total: 2,
          });
        }
        return HttpResponse.json({
          items: [
            { tool: "pmd", rule_id: "UnusedVariable", severity: "warning", file_path: "handler.go", line: 42, message: "Variable x unused", fingerprint: "fp1", status: "existing", source: "lint" },
            { tool: "spotbugs", rule_id: "NullPointer", severity: "error", file_path: "login.go", line: 18, message: "Possible null dereference", fingerprint: "fp3", status: "new", source: "lint" },
          ],
          total: 2,
        });
      }),
    );
  });

  it("should render diff summary with correct counts", async () => {
    renderDiffPage();
    expect(await screen.findByText("Run Diff")).toBeInTheDocument();
    expect(await screen.findByText("Added")).toBeInTheDocument();
    expect(await screen.findByText("Resolved")).toBeInTheDocument();
    expect(await screen.findByText("Unchanged")).toBeInTheDocument();

    const summaryCards = screen.getAllByText("1").filter(
      (el) => el.classList.contains("text-2xl"),
    );
    expect(summaryCards).toHaveLength(3);
  });

  it("should show added findings section", async () => {
    renderDiffPage();
    expect(await screen.findByText("Added findings")).toBeInTheDocument();
    expect(await screen.findByText("NullPointer")).toBeInTheDocument();
  });

  it("should show resolved findings section", async () => {
    renderDiffPage();
    expect(await screen.findByText("Resolved findings")).toBeInTheDocument();
    expect(await screen.findByText("EmptyCatch")).toBeInTheDocument();
  });

  it("should show unchanged findings", async () => {
    renderDiffPage();
    expect(await screen.findByText("Unchanged findings")).toBeInTheDocument();
    expect(await screen.findByText("UnusedVariable")).toBeInTheDocument();
  });
});
