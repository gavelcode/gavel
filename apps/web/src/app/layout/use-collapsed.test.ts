import { renderHook, act } from "@testing-library/react";
import { useCollapsed } from "./use-collapsed";

describe("useCollapsed", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("defaults to false when localStorage is empty", () => {
    const { result } = renderHook(() => useCollapsed());
    expect(result.current.collapsed).toBe(false);
  });

  it("reads initial state from localStorage", () => {
    localStorage.setItem("gavel-sidebar-collapsed", "true");
    const { result } = renderHook(() => useCollapsed());
    expect(result.current.collapsed).toBe(true);
  });

  it("toggles collapsed state and persists to localStorage", () => {
    const { result } = renderHook(() => useCollapsed());
    expect(result.current.collapsed).toBe(false);

    act(() => result.current.toggle());
    expect(result.current.collapsed).toBe(true);
    expect(localStorage.getItem("gavel-sidebar-collapsed")).toBe("true");

    act(() => result.current.toggle());
    expect(result.current.collapsed).toBe(false);
    expect(localStorage.getItem("gavel-sidebar-collapsed")).toBe("false");
  });
});
