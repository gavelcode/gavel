import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { KeyboardShortcutsOverlay } from "./keyboard-shortcuts";

describe("KeyboardShortcutsOverlay", () => {
  it("renders when open is true", () => {
    render(<KeyboardShortcutsOverlay open={true} onClose={() => {}} />);
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: /keyboard shortcuts/i })).toBeInTheDocument();
  });

  it("does not render when open is false", () => {
    render(<KeyboardShortcutsOverlay open={false} onClose={() => {}} />);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("shows shortcut entries", () => {
    render(<KeyboardShortcutsOverlay open={true} onClose={() => {}} />);
    expect(screen.getByText("j / k")).toBeInTheDocument();
    expect(screen.getByText("⌘K")).toBeInTheDocument();
    expect(screen.getByText("?")).toBeInTheDocument();
  });

  it("calls onClose when Escape is pressed", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(<KeyboardShortcutsOverlay open={true} onClose={onClose} />);
    await user.keyboard("{Escape}");
    expect(onClose).toHaveBeenCalled();
  });

  it("calls onClose when backdrop is clicked", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(<KeyboardShortcutsOverlay open={true} onClose={onClose} />);
    const backdrop = screen.getByTestId("shortcuts-backdrop");
    await user.click(backdrop);
    expect(onClose).toHaveBeenCalled();
  });
});
