import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderApp } from "@/test/render";
import { ProfilePage } from "./profile";

beforeEach(() => {
  localStorage.clear();
  document.documentElement.removeAttribute("data-density");
});

describe("ProfilePage density selector", () => {
  it("renders three density options", () => {
    renderApp(<ProfilePage />);
    expect(screen.getByRole("radio", { name: /comfortable/i })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /compact/i })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /dense/i })).toBeInTheDocument();
  });

  it("defaults to compact selected", () => {
    renderApp(<ProfilePage />);
    expect(screen.getByRole("radio", { name: /compact/i })).toBeChecked();
  });

  it("changes density on click", async () => {
    const user = userEvent.setup();
    renderApp(<ProfilePage />);
    await user.click(screen.getByRole("radio", { name: /comfortable/i }));
    expect(screen.getByRole("radio", { name: /comfortable/i })).toBeChecked();
    expect(document.documentElement.dataset.density).toBe("comfortable");
  });

  it("persists density to localStorage", async () => {
    const user = userEvent.setup();
    renderApp(<ProfilePage />);
    await user.click(screen.getByRole("radio", { name: /dense/i }));
    expect(localStorage.getItem("gavel-density")).toBe("dense");
  });
});
