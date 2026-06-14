import { screen } from "@testing-library/react";
import { renderApp } from "@/test/render";
import { CheckRow } from "./check-row";

describe("CheckRow", () => {
  it("renders pass status with check icon", () => {
    renderApp(<CheckRow label="No new blockers" status="pass" value="0 <= 0" />);
    expect(screen.getByText("No new blockers")).toBeInTheDocument();
    expect(screen.getByText("0 <= 0")).toBeInTheDocument();
  });

  it("renders fail status with X icon", () => {
    renderApp(<CheckRow label="Coverage on new code" status="fail" value="72% >= 80%" />);
    expect(screen.getByText("Coverage on new code")).toBeInTheDocument();
    expect(screen.getByText("72% >= 80%")).toBeInTheDocument();
  });

  it("renders warn status", () => {
    renderApp(<CheckRow label="Complexity" status="warn" value="15 <= 20" />);
    expect(screen.getByText("Complexity")).toBeInTheDocument();
  });
});
