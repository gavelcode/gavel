import { screen } from "@testing-library/react";
import { renderApp } from "@/test/render";
import { DesignSystemPage } from "./design-system";

describe("DesignSystemPage", () => {
  it("renders the page title", () => {
    renderApp(<DesignSystemPage />);
    expect(screen.getByRole("heading", { name: /design system/i })).toBeInTheDocument();
  });

  it("renders color palette section", () => {
    renderApp(<DesignSystemPage />);
    expect(screen.getByText(/color palette/i)).toBeInTheDocument();
    expect(screen.getByText("--primary")).toBeInTheDocument();
  });

  it("renders typography section", () => {
    renderApp(<DesignSystemPage />);
    expect(screen.getByText(/typography/i)).toBeInTheDocument();
  });

  it("renders button variants section", () => {
    renderApp(<DesignSystemPage />);
    expect(screen.getByRole("heading", { name: /button/i })).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: /default/i }).length).toBeGreaterThanOrEqual(1);
  });

  it("renders badge variants section", () => {
    renderApp(<DesignSystemPage />);
    expect(screen.getByText(/badges/i)).toBeInTheDocument();
  });

  it("renders spacing scale section", () => {
    renderApp(<DesignSystemPage />);
    expect(screen.getByText(/spacing/i)).toBeInTheDocument();
  });
});
