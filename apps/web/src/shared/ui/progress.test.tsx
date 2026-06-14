import { render, screen } from "@testing-library/react";
import { Progress } from "./progress";

describe("Progress", () => {
  it("should render a progressbar with correct value", () => {
    render(<Progress value={75} />);
    const bar = screen.getByRole("progressbar");
    expect(bar).toHaveAttribute("aria-valuenow", "75");
    expect(bar).toHaveAttribute("aria-valuemin", "0");
    expect(bar).toHaveAttribute("aria-valuemax", "100");
  });

  it("should use h-2 height by default", () => {
    const { container } = render(<Progress value={50} />);
    const track = container.querySelector("[role='progressbar']");
    expect(track?.className).toContain("h-2");
  });

  it("should not show label by default", () => {
    render(<Progress value={80} />);
    expect(screen.queryByText("80%")).not.toBeInTheDocument();
  });

  it("should show percentage label when showLabel is true", () => {
    render(<Progress value={80} showLabel />);
    expect(screen.getByText("80%")).toBeInTheDocument();
  });

  it("should apply success color for values >= 80", () => {
    const { container } = render(<Progress value={85} />);
    const fill = container.querySelector("[role='progressbar'] > div");
    expect(fill?.className).toContain("bg-success");
  });

  it("should apply warning color for values >= 60 and < 80", () => {
    const { container } = render(<Progress value={65} />);
    const fill = container.querySelector("[role='progressbar'] > div");
    expect(fill?.className).toContain("bg-warning");
  });

  it("should apply danger color for values < 60", () => {
    const { container } = render(<Progress value={40} />);
    const fill = container.querySelector("[role='progressbar'] > div");
    expect(fill?.className).toContain("bg-danger");
  });

  it("should clamp value between 0 and 100", () => {
    const { container } = render(<Progress value={150} />);
    const fill = container.querySelector("[role='progressbar'] > div") as HTMLElement;
    expect(fill?.style.width).toBe("100%");
  });
});
