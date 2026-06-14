import { screen } from "@testing-library/react";
import { renderApp } from "@/test/render";
import { QualityGateBanner } from "./quality-gate-banner";

describe("QualityGateBanner", () => {
  it("renders pass status with title and description", () => {
    renderApp(
      <QualityGateBanner status="pass" title="Quality gate passed" description="All conditions met." />,
    );
    expect(screen.getByText("Quality gate passed")).toBeInTheDocument();
    expect(screen.getByText("All conditions met.")).toBeInTheDocument();
  });

  it("renders fail status", () => {
    renderApp(
      <QualityGateBanner status="fail" title="Quality gate failed" description="2 of 3 conditions met." />,
    );
    expect(screen.getByText("Quality gate failed")).toBeInTheDocument();
    expect(screen.getByText("2 of 3 conditions met.")).toBeInTheDocument();
  });

  it("renders warn status", () => {
    renderApp(
      <QualityGateBanner status="warn" title="Warning" description="Coverage is low." />,
    );
    expect(screen.getByText("Warning")).toBeInTheDocument();
  });

  it("renders stats when provided", () => {
    renderApp(
      <QualityGateBanner
        status="pass"
        title="Passed"
        description="OK"
        stats={[
          { label: "Findings", value: "0", tone: "success" },
          { label: "Coverage", value: "92%", tone: "warning" },
          { label: "Score", value: "8.5", tone: "danger" },
          { label: "Total", value: "42" },
        ]}
      />,
    );
    expect(screen.getByText("0")).toBeInTheDocument();
    expect(screen.getByText("Findings")).toBeInTheDocument();
    expect(screen.getByText("92%")).toBeInTheDocument();
    expect(screen.getByText("Coverage")).toBeInTheDocument();
    expect(screen.getByText("8.5")).toBeInTheDocument();
    expect(screen.getByText("42")).toBeInTheDocument();
  });

  it("does not render stats section when not provided", () => {
    const { container } = renderApp(
      <QualityGateBanner status="pass" title="OK" description="Fine" />,
    );
    expect(container.querySelector(".font-mono")).not.toBeInTheDocument();
  });
});
