import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TokenReveal } from "./token-reveal";
import type { CreateTokenResult } from "../model";

const result: CreateTokenResult = {
  id: "tok-1",
  name: "ci-token",
  scopes: ["ingest"],
  token: "gav_abc123",
  prefix: "gav_abc",
};

describe("TokenReveal", () => {
  it("copies token to clipboard on click", async () => {
    const user = userEvent.setup();

    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", {
      value: { writeText, readText: vi.fn(), read: vi.fn(), write: vi.fn(), addEventListener: vi.fn(), removeEventListener: vi.fn(), dispatchEvent: vi.fn() },
      writable: true,
      configurable: true,
    });

    render(<TokenReveal result={result} onDismiss={vi.fn()} />);

    const copyButton = screen.getAllByRole("button")[0];
    await user.click(copyButton);

    await waitFor(() => {
      expect(writeText).toHaveBeenCalledWith("gav_abc123");
    });
  });

  it("calls onDismiss when dismiss button is clicked", async () => {
    const onDismiss = vi.fn();
    const user = userEvent.setup();
    render(<TokenReveal result={result} onDismiss={onDismiss} />);

    await user.click(screen.getByRole("button", { name: /dismiss/i }));

    expect(onDismiss).toHaveBeenCalledOnce();
  });

  it("displays the token text", () => {
    render(<TokenReveal result={result} onDismiss={vi.fn()} />);

    expect(screen.getByText("gav_abc123")).toBeInTheDocument();
  });
});
