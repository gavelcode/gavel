import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ThemeProvider } from "@/shared/lib/theme";
import { ThemeSelector } from "./theme-selector";

function renderWithTheme() {
  return render(
    <ThemeProvider>
      <ThemeSelector />
    </ThemeProvider>,
  );
}

beforeEach(() => {
  localStorage.clear();
  document.documentElement.classList.remove("dark");
});

describe("ThemeSelector", () => {
  it("renders three theme options", () => {
    renderWithTheme();
    expect(screen.getByRole("radio", { name: /auto/i })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /light/i })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /dark/i })).toBeInTheDocument();
  });

  it("defaults to system/auto selected", () => {
    renderWithTheme();
    expect(screen.getByRole("radio", { name: /auto/i })).toBeChecked();
  });

  it("switches to light on click", async () => {
    const user = userEvent.setup();
    renderWithTheme();
    await user.click(screen.getByRole("radio", { name: /light/i }));
    expect(screen.getByRole("radio", { name: /light/i })).toBeChecked();
    expect(localStorage.getItem("gavel-theme")).toBe("light");
  });

  it("switches to dark on click", async () => {
    const user = userEvent.setup();
    renderWithTheme();
    await user.click(screen.getByRole("radio", { name: /dark/i }));
    expect(screen.getByRole("radio", { name: /dark/i })).toBeChecked();
    expect(localStorage.getItem("gavel-theme")).toBe("dark");
  });

  it("has radiogroup role", () => {
    renderWithTheme();
    expect(screen.getByRole("radiogroup", { name: /theme/i })).toBeInTheDocument();
  });
});
