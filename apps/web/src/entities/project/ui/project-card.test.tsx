import { screen } from "@testing-library/react";
import { renderApp } from "@/test/render";
import { ProjectCard } from "./project-card";
import type { ProjectSummary } from "../model";

describe("ProjectCard", () => {
  const project: ProjectSummary = {
    id: "proj-aaa-111",
    key: "//services/payment/...",
    name: "Payment",
    defaultBranch: "main",
    latestVerdict: "pass",
    totalFindings: 12,
    createdAt: "2025-01-01T00:00:00Z",
  };

  it("renders project name and findings count", () => {
    renderApp(<ProjectCard project={project} />);
    expect(screen.getByText("Payment")).toBeInTheDocument();
    expect(screen.getByText("12 findings")).toBeInTheDocument();
  });

  it("renders verdict badge", () => {
    renderApp(<ProjectCard project={project} />);
    expect(screen.getByText("Passed")).toBeInTheDocument();
  });
});
