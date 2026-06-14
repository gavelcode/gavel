import { screen, within } from "@testing-library/react";
import { Routes, Route } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { renderApp } from "@/test/render";
import { server } from "@/test/msw-server";
import { GavelspaceDetailPage } from "../gavelspace-detail";
import { FindingDetailPage } from "./finding-detail";

function renderFindingDetail(route: string) {
  return renderApp(
    <Routes>
      <Route path="/gavelspaces/:name" element={<GavelspaceDetailPage />}>
        <Route path="findings/:findingId" element={<FindingDetailPage />} />
      </Route>
    </Routes>,
    { route },
  );
}

describe("FindingDetailPage (Phase 25)", () => {
  beforeEach(() => {
    server.use(
      http.get("/api/v1/projects/:key/source", ({ params, request }) => {
        const url = new URL(request.url);
        if (
          params.key === "alpha-proj" &&
          url.searchParams.get("commit") === "abc1234" &&
          url.searchParams.get("path") === "services/payment/handler.go"
        ) {
          const body =
            "package payment\n\nfunc Handle() {\n  // ...\n}\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\nvar x = 1 // line 42 marker\n";
          return new HttpResponse(body, {
            status: 200,
            headers: { "Content-Type": "text/plain; charset=utf-8" },
          });
        }
        return new HttpResponse("not found", { status: 404 });
      }),
    );
  });

  it("renders the viewer for the matching finding when deep-linked cold", async () => {
    renderFindingDetail("/gavelspaces/alpha/findings/fp1");

    await screen.findByTestId("code-viewer");
    const viewer = screen.getByTestId("code-viewer");
    const activeRow = viewer.querySelector("[data-active='true']") as HTMLElement;
    expect(activeRow).not.toBeNull();
    expect(activeRow).toHaveAttribute("data-line-number", "42");
    expect(within(viewer).getAllByTestId("finding-marker").length).toBeGreaterThan(0);
  });

  it("shows a friendly source-missing state when the source endpoint returns 404", async () => {
    server.use(
      http.get("/api/v1/projects/:key/source", () =>
        new HttpResponse("not found", { status: 404 }),
      ),
    );

    renderFindingDetail("/gavelspaces/alpha/findings/fp1");

    expect(
      await screen.findByText(/source not available for this commit/i),
    ).toBeInTheDocument();
    expect(screen.queryByTestId("code-viewer")).not.toBeInTheDocument();
  });

  it("shows a finding-not-found state when the fingerprint matches no finding in the gavelspace", async () => {
    renderFindingDetail("/gavelspaces/alpha/findings/does-not-exist");

    expect(await screen.findByText(/finding not found/i)).toBeInTheDocument();
    expect(screen.queryByTestId("code-viewer")).not.toBeInTheDocument();
  });

  it("renders a back-to-findings link that points to the findings tab", async () => {
    renderFindingDetail("/gavelspaces/alpha/findings/fp1");

    const link = await screen.findByRole("link", { name: /back to findings/i });
    expect(link).toHaveAttribute("href", "/gavelspaces/alpha/findings");
  });

  it("uses fetchSourceWithContext when finding has casefileId and scrolls to active line", async () => {
    const scrollIntoViewMock = vi.fn();
    HTMLElement.prototype.scrollIntoView = scrollIntoViewMock;

    server.use(
      http.get("/api/v1/findings", ({ request }) => {
        const url = new URL(request.url);
        const gavelspace = url.searchParams.get("gavelspace");
        if (gavelspace === "alpha") {
          return HttpResponse.json({
            items: [
              {
                tool: "pmd",
                rule_id: "UnusedVariable",
                severity: "warning",
                file_path: "services/payment/handler.go",
                line: 42,
                message: "Variable 'x' is never used",
                fingerprint: "fp1",
                status: "new",
                source: "lint",
                commit_sha: "abc1234",
                project_key: "alpha-proj",
                casefile_id: "cf-aaa-111",
              },
            ],
            total: 1,
          });
        }
        return HttpResponse.json({ items: [], total: 0 });
      }),
      http.get("/api/v1/projects/:key/source", ({ params, request }) => {
        const url = new URL(request.url);
        if (
          params.key === "alpha-proj" &&
          url.searchParams.get("casefile") === "cf-aaa-111"
        ) {
          return HttpResponse.json({
            content:
              "package payment\n\nfunc Handle() {\n  // ...\n}\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\n// ...\n\nvar x = 1 // line 42 marker\n",
            coverage: {
              covered_lines: [1, 2, 3],
              uncovered_lines: [42],
            },
          });
        }
        return new HttpResponse("not found", { status: 404 });
      }),
    );

    renderFindingDetail("/gavelspaces/alpha/findings/fp1");

    await screen.findByTestId("code-viewer");
    const viewer = screen.getByTestId("code-viewer");
    const activeRow = viewer.querySelector("[data-active='true']") as HTMLElement;
    expect(activeRow).not.toBeNull();
    expect(scrollIntoViewMock).toHaveBeenCalled();
  });
});
