import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { renderApp } from "@/test/render";
import { server } from "@/test/msw-server";
import { AppLayout } from "@/app/layout/sidebar";
import { Routes, Route } from "react-router-dom";

function renderWithRoutes() {
  return renderApp(
    <Routes>
      <Route element={<AppLayout />}>
        <Route path="/" element={<div>Home</div>} />
        <Route path="/projects/:id" element={<div>Project Detail</div>} />
        <Route path="/casefiles/:id" element={<div>CaseFile Detail</div>} />
      </Route>
    </Routes>,
  );
}

describe("CommandPalette", () => {
  beforeEach(() => {
    server.use(
      http.get("/api/v1/search", ({ request }) => {
        const url = new URL(request.url);
        const q = (url.searchParams.get("q") ?? "").toLowerCase();
        const all = [
          { type: "project", id: "proj-aaa-111", title: "Payment", subtitle: "//services/payment/...", url: "/projects/proj-aaa-111" },
          { type: "project", id: "proj-bbb-222", title: "Auth", subtitle: "//services/auth/...", url: "/projects/proj-bbb-222" },
          { type: "finding", id: "fp1", title: "UnusedVariable", subtitle: "services/payment/handler.go:42", url: "/casefiles/cf-aaa-111" },
          { type: "finding", id: "fp2", title: "NP_NULL_ON_SOME_PATH", subtitle: "services/auth/login.go:18", url: "/casefiles/cf-bbb-222" },
          { type: "casefile", id: "cf-aaa-111", title: "abc1234", subtitle: "main · 3 findings", url: "/casefiles/cf-aaa-111" },
          { type: "casefile", id: "cf-bbb-222", title: "def5678", subtitle: "main · 5 findings", url: "/casefiles/cf-bbb-222" },
        ];
        const results = q ? all.filter((r) => r.title.toLowerCase().includes(q) || r.subtitle.toLowerCase().includes(q)) : [];
        return HttpResponse.json({ results });
      }),
    );
  });

  it("should open search modal when pressing ⌘K", async () => {
    const user = userEvent.setup();
    renderWithRoutes();
    await user.keyboard("{Meta>}k{/Meta}");
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/search/i)).toHaveFocus();
  });

  it("should open search modal when pressing Ctrl+K", async () => {
    const user = userEvent.setup();
    renderWithRoutes();
    await user.keyboard("{Control>}k{/Control}");
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });

  it("should close search modal on Escape", async () => {
    const user = userEvent.setup();
    renderWithRoutes();
    await user.keyboard("{Meta>}k{/Meta}");
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("should show results after typing with debounce", async () => {
    const user = userEvent.setup();
    renderWithRoutes();
    await user.keyboard("{Meta>}k{/Meta}");
    await user.type(screen.getByPlaceholderText(/search/i), "Payment");
    expect(await screen.findByText("Payment")).toBeInTheDocument();
  });

  it("should navigate to result URL on click", async () => {
    const user = userEvent.setup();
    renderWithRoutes();
    await user.keyboard("{Meta>}k{/Meta}");
    await user.type(screen.getByPlaceholderText(/search/i), "Payment");
    const result = await screen.findByText("Payment");
    await user.click(result.closest("button")!);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(screen.getByText("Project Detail")).toBeInTheDocument();
  });

  it("should navigate to result URL on Enter", async () => {
    const user = userEvent.setup();
    renderWithRoutes();
    await user.keyboard("{Meta>}k{/Meta}");
    await user.type(screen.getByPlaceholderText(/search/i), "Payment");
    await screen.findByText("Payment");
    await user.keyboard("{Enter}");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(screen.getByText("Project Detail")).toBeInTheDocument();
  });

  it("should show empty state when no results match", async () => {
    const user = userEvent.setup();
    renderWithRoutes();
    await user.keyboard("{Meta>}k{/Meta}");
    await user.type(screen.getByPlaceholderText(/search/i), "zzzznotfound");
    expect(await screen.findByText(/no results for/i)).toBeInTheDocument();
  });

  it("should group results by type (project, finding, casefile)", async () => {
    const user = userEvent.setup();
    renderWithRoutes();
    await user.keyboard("{Meta>}k{/Meta}");
    await user.type(screen.getByPlaceholderText(/search/i), "a");

    await screen.findByText("Payment");
    const dialog = screen.getByRole("dialog");
    expect(within(dialog).getByText("Projects")).toBeInTheDocument();
    expect(within(dialog).getByText("Findings")).toBeInTheDocument();
    expect(within(dialog).getByText("Case Files")).toBeInTheDocument();
  });

  it("should trap focus within dialog on Tab", async () => {
    const user = userEvent.setup();
    renderWithRoutes();
    await user.keyboard("{Meta>}k{/Meta}");
    const dialog = screen.getByRole("dialog");
    const input = screen.getByPlaceholderText(/search/i);
    expect(input).toHaveFocus();
    await user.tab();
    expect(dialog.contains(document.activeElement)).toBe(true);
  });

  it("should trap focus within dialog on Shift+Tab", async () => {
    const user = userEvent.setup();
    renderWithRoutes();
    await user.keyboard("{Meta>}k{/Meta}");
    const dialog = screen.getByRole("dialog");
    expect(screen.getByPlaceholderText(/search/i)).toHaveFocus();
    await user.tab({ shift: true });
    expect(dialog.contains(document.activeElement)).toBe(true);
  });

  it("should have aria-label on dialog", async () => {
    const user = userEvent.setup();
    renderWithRoutes();
    await user.keyboard("{Meta>}k{/Meta}");
    expect(screen.getByRole("dialog", { name: /search/i })).toBeInTheDocument();
  });

  it("should have entry animation classes on the overlay and panel", async () => {
    const user = userEvent.setup();
    renderWithRoutes();
    await user.keyboard("{Meta>}k{/Meta}");
    const dialog = screen.getByRole("dialog");
    expect(dialog.querySelector("[data-testid='palette-overlay']")?.className).toMatch(/animate-fade-in/);
    expect(dialog.querySelector("[data-testid='palette-panel']")?.className).toMatch(/animate-scale-in/);
  });

  it("should move selection down with ArrowDown key", async () => {
    const user = userEvent.setup();
    renderWithRoutes();
    await user.keyboard("{Meta>}k{/Meta}");
    await user.type(screen.getByPlaceholderText(/search/i), "a");
    await screen.findByText("Payment");
    const results = screen.getAllByTestId("search-result");
    expect(results[0].className).toContain("bg-primary/10");
    await user.keyboard("{ArrowDown}");
    const updatedResults = screen.getAllByTestId("search-result");
    expect(updatedResults[1].className).toContain("bg-primary/10");
    expect(updatedResults[0].className).not.toContain("bg-primary/10");
  });

  it("should move selection up with ArrowUp key", async () => {
    const user = userEvent.setup();
    renderWithRoutes();
    await user.keyboard("{Meta>}k{/Meta}");
    await user.type(screen.getByPlaceholderText(/search/i), "a");
    await screen.findByText("Payment");
    await user.keyboard("{ArrowDown}");
    await user.keyboard("{ArrowDown}");
    const afterDown = screen.getAllByTestId("search-result");
    expect(afterDown[2].className).toContain("bg-primary/10");
    await user.keyboard("{ArrowUp}");
    const afterUp = screen.getAllByTestId("search-result");
    expect(afterUp[1].className).toContain("bg-primary/10");
    expect(afterUp[2].className).not.toContain("bg-primary/10");
  });
});
