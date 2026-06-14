import { render } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Card } from "./card";

describe("Card", () => {
  it("should not have shadow by default", () => {
    const { container } = render(<Card>Content</Card>);
    const card = container.firstElementChild!;
    expect(card.className).not.toContain("shadow-raised");
  });

  it("should have shadow when elevated", () => {
    const { container } = render(<Card elevated>Content</Card>);
    const card = container.firstElementChild!;
    expect(card.className).toContain("shadow-raised");
  });

  it("should always have border and rounded corners", () => {
    const { container } = render(<Card>Content</Card>);
    const card = container.firstElementChild!;
    expect(card.className).toContain("border");
    expect(card.className).toContain("rounded-xl");
  });

  it("should have tabIndex=0 when onClick is provided", () => {
    const { container } = render(<Card onClick={() => {}}>Content</Card>);
    const card = container.firstElementChild!;
    expect(card).toHaveAttribute("tabindex", "0");
  });

  it("should not have tabIndex when onClick is absent", () => {
    const { container } = render(<Card>Content</Card>);
    const card = container.firstElementChild!;
    expect(card).not.toHaveAttribute("tabindex");
  });

  it("should activate on Enter key when onClick is provided", async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    const { container } = render(<Card onClick={onClick}>Content</Card>);
    const card = container.firstElementChild!;
    (card as HTMLElement).focus();
    await user.keyboard("{Enter}");
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it("should activate on Space key when onClick is provided", async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    const { container } = render(<Card onClick={onClick}>Content</Card>);
    const card = container.firstElementChild!;
    (card as HTMLElement).focus();
    await user.keyboard(" ");
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it("should have focus-visible ring when onClick is provided", () => {
    const { container } = render(<Card onClick={() => {}}>Content</Card>);
    const card = container.firstElementChild!;
    expect(card.className).toContain("focus-visible:ring-1");
  });

  it("should have hover lift when elevated", () => {
    const { container } = render(<Card elevated>Content</Card>);
    const card = container.firstElementChild!;
    expect(card.className).toMatch(/hover:-translate-y/);
  });

  it("should not have hover lift when not elevated", () => {
    const { container } = render(<Card>Content</Card>);
    const card = container.firstElementChild!;
    expect(card.className).not.toMatch(/hover:-translate-y/);
  });
});
