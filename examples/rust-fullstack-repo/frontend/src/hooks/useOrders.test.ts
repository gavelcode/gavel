import { renderHook, act } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { useOrders } from "./useOrders";

describe("useOrders", () => {
  it("starts with empty orders", () => {
    const { result } = renderHook(() => useOrders());
    expect(result.current.orders).toEqual([]);
    expect(result.current.loading).toBe(false);
  });

  it("addOrder appends to list", () => {
    const { result } = renderHook(() => useOrders());
    act(() => {
      result.current.addOrder({ id: "O1", total: 25.0, status: "confirmed" });
    });
    expect(result.current.orders).toHaveLength(1);
    expect(result.current.orders[0].id).toBe("O1");
  });

  it("addOrder preserves existing orders", () => {
    const { result } = renderHook(() => useOrders());
    act(() => {
      result.current.addOrder({ id: "O1", total: 10.0, status: "draft" });
      result.current.addOrder({ id: "O2", total: 20.0, status: "shipped" });
    });
    expect(result.current.orders).toHaveLength(2);
  });
});
