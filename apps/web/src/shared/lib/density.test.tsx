import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DensityProvider } from "./density";
import { useDensity } from "./use-density";

function TestConsumer() {
  const { density, setDensity } = useDensity();
  return (
    <div>
      <span data-testid="current">{density}</span>
      <button onClick={() => setDensity("comfortable")}>comfortable</button>
      <button onClick={() => setDensity("compact")}>compact</button>
      <button onClick={() => setDensity("dense")}>dense</button>
    </div>
  );
}

beforeEach(() => {
  localStorage.clear();
  document.documentElement.removeAttribute("data-density");
});

describe("DensityProvider", () => {
  it("defaults to compact", () => {
    render(
      <DensityProvider>
        <TestConsumer />
      </DensityProvider>,
    );
    expect(screen.getByTestId("current").textContent).toBe("compact");
    expect(document.documentElement.dataset.density).toBe("compact");
  });

  it("applies data-density attribute to html element", async () => {
    const user = userEvent.setup();
    render(
      <DensityProvider>
        <TestConsumer />
      </DensityProvider>,
    );
    await user.click(screen.getByText("comfortable"));
    expect(document.documentElement.dataset.density).toBe("comfortable");
    expect(screen.getByTestId("current").textContent).toBe("comfortable");
  });

  it("persists to localStorage", async () => {
    const user = userEvent.setup();
    render(
      <DensityProvider>
        <TestConsumer />
      </DensityProvider>,
    );
    await user.click(screen.getByText("dense"));
    expect(localStorage.getItem("gavel-density")).toBe("dense");
  });

  it("restores from localStorage", () => {
    localStorage.setItem("gavel-density", "comfortable");
    render(
      <DensityProvider>
        <TestConsumer />
      </DensityProvider>,
    );
    expect(screen.getByTestId("current").textContent).toBe("comfortable");
    expect(document.documentElement.dataset.density).toBe("comfortable");
  });
});
