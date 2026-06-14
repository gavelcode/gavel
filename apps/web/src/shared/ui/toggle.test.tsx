import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Toggle } from "./toggle";

describe("Toggle", () => {
  it("should render as a switch with aria-checked", () => {
    render(<Toggle checked={false} />);
    const toggle = screen.getByRole("switch");
    expect(toggle).toHaveAttribute("aria-checked", "false");
  });

  it("should have focus-visible ring styling", () => {
    render(<Toggle checked={false} />);
    const toggle = screen.getByRole("switch");
    expect(toggle.className).toContain("focus-visible:ring-1");
  });

  it("should toggle on click", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<Toggle checked={false} onChange={onChange} />);
    await user.click(screen.getByRole("switch"));
    expect(onChange).toHaveBeenCalledWith(true);
  });

  it("should use motion token duration for knob transition", () => {
    render(<Toggle checked={false} />);
    const knob = screen.getByRole("switch").querySelector("span")!;
    expect(knob.className).toMatch(/duration-normal/);
  });
});
