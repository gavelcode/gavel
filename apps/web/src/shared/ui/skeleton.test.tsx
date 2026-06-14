import { render, screen } from "@testing-library/react";
import { Skeleton, SkeletonText, SkeletonCard } from "./skeleton";

describe("Skeleton", () => {
  it("should render with shimmer animation", () => {
    render(<Skeleton data-testid="skel" />);
    const el = screen.getByTestId("skel");
    expect(el.className).toMatch(/animate-shimmer/);
    expect(el.className).toMatch(/bg-muted/);
  });

  it("should accept custom className and dimensions", () => {
    render(<Skeleton data-testid="skel" className="h-8 w-32" />);
    const el = screen.getByTestId("skel");
    expect(el.className).toMatch(/h-8/);
    expect(el.className).toMatch(/w-32/);
  });

  it("should have aria-hidden for accessibility", () => {
    render(<Skeleton data-testid="skel" />);
    expect(screen.getByTestId("skel")).toHaveAttribute("aria-hidden", "true");
  });
});

describe("SkeletonText", () => {
  it("should render the specified number of lines", () => {
    render(<SkeletonText lines={3} data-testid="skel-text" />);
    const container = screen.getByTestId("skel-text");
    const lines = container.querySelectorAll("[aria-hidden='true']");
    expect(lines).toHaveLength(3);
  });

  it("should make the last line shorter", () => {
    render(<SkeletonText lines={2} data-testid="skel-text" />);
    const container = screen.getByTestId("skel-text");
    const lines = container.querySelectorAll("[aria-hidden='true']");
    expect(lines[1].className).toMatch(/w-2\/3/);
  });
});

describe("SkeletonCard", () => {
  it("should render a card-shaped skeleton with header and body lines", () => {
    render(<SkeletonCard data-testid="skel-card" />);
    const container = screen.getByTestId("skel-card");
    const skeletons = container.querySelectorAll("[aria-hidden='true']");
    expect(skeletons.length).toBeGreaterThanOrEqual(3);
  });
});
