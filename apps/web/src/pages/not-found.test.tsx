import { screen } from "@testing-library/react";
import { renderApp } from "@/test/render";
import { NotFoundPage } from "./not-found";

describe("NotFoundPage", () => {
  it("renders 404 message with back link", () => {
    renderApp(<NotFoundPage />);
    expect(screen.getByText("Page not found")).toBeInTheDocument();
    expect(screen.getByText("Back to Projects")).toBeInTheDocument();
  });
});
