import { render, screen } from "@testing-library/react";
import { Button } from "./button";

describe("Button", () => {
  it("should render children as label", () => {
    render(<Button>Save</Button>);
    expect(screen.getByRole("button", { name: "Save" })).toBeInTheDocument();
  });

  it("should be disabled when loading", () => {
    render(<Button loading>Save</Button>);
    expect(screen.getByRole("button")).toBeDisabled();
  });

  it("should show spinner when loading", () => {
    render(<Button loading>Save</Button>);
    expect(screen.getByRole("button").querySelector("svg.animate-spin")).toBeInTheDocument();
  });

  it("should keep label visible when loading", () => {
    render(<Button loading>Save</Button>);
    expect(screen.getByRole("button", { name: /save/i })).toBeInTheDocument();
  });

  it("should apply success variant styles", () => {
    render(<Button variant="success">Confirm</Button>);
    const btn = screen.getByRole("button", { name: "Confirm" });
    expect(btn.className).toContain("bg-success");
  });

  it("should apply default variant when no variant specified", () => {
    render(<Button>Click</Button>);
    const btn = screen.getByRole("button", { name: "Click" });
    expect(btn.className).toContain("bg-primary");
  });

  it("should have active:scale press feedback", () => {
    render(<Button>Press</Button>);
    const btn = screen.getByRole("button", { name: "Press" });
    expect(btn.className).toMatch(/active:scale/);
  });
});
