import { render, screen } from "@testing-library/react";
import { Severity } from "./severity";

describe("Severity", () => {
  it("should render label text for known levels", () => {
    render(<Severity level="error" />);
    expect(screen.getByText("Error")).toBeInTheDocument();
  });

  it("should render circular dot indicator", () => {
    const { container } = render(<Severity level="warning" />);
    const dot = container.querySelector("span.rounded-full");
    expect(dot).toBeInTheDocument();
  });

  it("should not render square dot", () => {
    const { container } = render(<Severity level="error" />);
    expect(container.querySelector("span.rounded-sm")).not.toBeInTheDocument();
  });

  it("should have aria-label on the dot for screen readers", () => {
    render(<Severity level="error" />);
    expect(screen.getByLabelText("Error")).toBeInTheDocument();
  });

  it("should render fallback for unknown level", () => {
    render(<Severity level="unknown-thing" />);
    expect(screen.getByText("Unknown")).toBeInTheDocument();
  });
});
