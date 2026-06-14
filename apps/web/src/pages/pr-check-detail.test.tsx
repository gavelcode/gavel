import { screen } from "@testing-library/react";
import { Routes, Route } from "react-router-dom";
import { renderApp } from "@/test/render";
import "@/test/msw-server";
import { PRCheckDetailPage } from "./pr-check-detail";

function renderDetail(id = "pr-aaa-111") {
  return renderApp(
    <Routes>
      <Route path="/pr-checks/:id" element={<PRCheckDetailPage />} />
    </Routes>,
    { route: `/pr-checks/${id}` },
  );
}

describe("PRCheckDetailPage", () => {
  it("shows loading skeleton before data arrives", () => {
    renderDetail();
    expect(screen.getByText("...")).toBeInTheDocument();
  });

  it("renders PR details after loading", async () => {
    renderDetail();
    expect(await screen.findByText("Use parameterized queries in posting pipeline")).toBeInTheDocument();
    expect(screen.getByText("Marek Novák")).toBeInTheDocument();
    expect(screen.getByText("fix/posting-sql")).toBeInTheDocument();
  });

  it("shows quality gate banner with pass status", async () => {
    renderDetail();
    await screen.findByText("Use parameterized queries in posting pipeline");
    expect(screen.getByText("Quality gate passed")).toBeInTheDocument();
    expect(screen.getByText("2 of 2 conditions met.")).toBeInTheDocument();
  });

  it("renders quality gate conditions as check rows", async () => {
    renderDetail();
    await screen.findByText("Use parameterized queries in posting pipeline");
    expect(screen.getByText("No new blocker issues")).toBeInTheDocument();
    expect(screen.getByText("Coverage on new code")).toBeInTheDocument();
  });

  it("shows commit SHA truncated to 7 chars", async () => {
    renderDetail();
    await screen.findByText("Use parameterized queries in posting pipeline");
    expect(screen.getByText("e8c1d2f")).toBeInTheDocument();
  });
});
