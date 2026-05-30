import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import { describe, it, expect, vi, afterEach } from "vitest";
import { OrderForm } from "./OrderForm";

describe("OrderForm", () => {
  afterEach(() => cleanup());

  it("renders input fields and submit button", () => {
    render(<OrderForm onSubmit={() => {}} isLoggedIn={false} />);
    expect(screen.getByPlaceholderText("Product ID")).toBeDefined();
    expect(screen.getByRole("button", { name: "Add to Order" })).toBeDefined();
  });

  it("calls onSubmit with product and quantity", () => {
    const onSubmit = vi.fn();
    render(<OrderForm onSubmit={onSubmit} isLoggedIn={false} />);

    fireEvent.change(screen.getByPlaceholderText("Product ID"), {
      target: { value: "P1" },
    });
    fireEvent.change(screen.getByRole("spinbutton"), {
      target: { value: "3" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Add to Order" }));

    expect(onSubmit).toHaveBeenCalledWith("P1", 3);
  });

  it("defaults quantity to 1", () => {
    const onSubmit = vi.fn();
    render(<OrderForm onSubmit={onSubmit} isLoggedIn={false} />);

    fireEvent.change(screen.getByPlaceholderText("Product ID"), {
      target: { value: "P2" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Add to Order" }));

    expect(onSubmit).toHaveBeenCalledWith("P2", 1);
  });
});
