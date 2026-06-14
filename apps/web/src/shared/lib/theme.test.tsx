import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useContext } from "react";
import { ThemeProvider } from "./theme";
import { ThemeContext } from "./theme-context";

function TestConsumer() {
  const { resolved, highContrast, toggleTheme, setHighContrast } = useContext(ThemeContext);
  return (
    <div>
      <span data-testid="resolved">{resolved}</span>
      <span data-testid="hc">{String(highContrast)}</span>
      <button onClick={toggleTheme}>toggle</button>
      <button onClick={() => setHighContrast(true)}>hc-on</button>
      <button onClick={() => setHighContrast(false)}>hc-off</button>
    </div>
  );
}

describe("ThemeProvider", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("toggleTheme switches from light to dark and updates localStorage", async () => {
    const user = userEvent.setup();

    render(
      <ThemeProvider>
        <TestConsumer />
      </ThemeProvider>,
    );

    expect(screen.getByTestId("resolved").textContent).toBe("light");

    await user.click(screen.getByRole("button", { name: "toggle" }));

    expect(screen.getByTestId("resolved").textContent).toBe("dark");
    expect(localStorage.getItem("gavel-theme")).toBe("dark");
  });

  it("toggleTheme switches from dark back to light", async () => {
    localStorage.setItem("gavel-theme", "dark");
    const user = userEvent.setup();

    render(
      <ThemeProvider>
        <TestConsumer />
      </ThemeProvider>,
    );

    expect(screen.getByTestId("resolved").textContent).toBe("dark");

    await user.click(screen.getByRole("button", { name: "toggle" }));

    expect(screen.getByTestId("resolved").textContent).toBe("light");
    expect(localStorage.getItem("gavel-theme")).toBe("light");
  });

  it("setHighContrast(true) updates state and localStorage", async () => {
    const user = userEvent.setup();

    render(
      <ThemeProvider>
        <TestConsumer />
      </ThemeProvider>,
    );

    expect(screen.getByTestId("hc").textContent).toBe("false");

    await user.click(screen.getByRole("button", { name: "hc-on" }));

    expect(screen.getByTestId("hc").textContent).toBe("true");
    expect(localStorage.getItem("gavel-high-contrast")).toBe("true");
  });

  it("setHighContrast(false) clears the flag", async () => {
    localStorage.setItem("gavel-high-contrast", "true");
    const user = userEvent.setup();

    render(
      <ThemeProvider>
        <TestConsumer />
      </ThemeProvider>,
    );

    expect(screen.getByTestId("hc").textContent).toBe("true");

    await user.click(screen.getByRole("button", { name: "hc-off" }));

    expect(screen.getByTestId("hc").textContent).toBe("false");
    expect(localStorage.getItem("gavel-high-contrast")).toBe("false");
  });
});
