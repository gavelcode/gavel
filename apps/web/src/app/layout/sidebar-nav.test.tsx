import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";
import { renderApp } from "@/test/render";
import "@/test/msw-server";
import { AppLayout } from "./sidebar";

describe("SidebarNav", () => {
  it("should contain only Gavelspaces and API Tokens in nav (yaml-canonical: removed top-level Projects/Issues/PR Checks/Case Files)", () => {
    renderApp(<AppLayout />, { auth: { user: { id: 2, email: "viewer@local", displayName: "Viewer", role: "viewer", mustChangePassword: false } } });

    expect(screen.getByRole("link", { name: /gavelspaces/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /api tokens/i })).toBeInTheDocument();

    expect(screen.queryByRole("link", { name: /^projects$/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /^issues$/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /pr checks/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /case files/i })).not.toBeInTheDocument();
  });

  it("should highlight active nav item based on current route", () => {
    renderApp(<AppLayout />, { route: "/gavelspaces" });
    const gsLink = screen.getByRole("link", { name: /gavelspaces/i });
    expect(gsLink).toHaveClass("bg-muted");
  });

  it("should navigate when clicking nav item", async () => {
    const user = userEvent.setup();
    renderApp(<AppLayout />, { route: "/" });
    await user.click(screen.getByRole("link", { name: /gavelspaces/i }));
    const gsLink = screen.getByRole("link", { name: /gavelspaces/i });
    expect(gsLink).toHaveClass("bg-muted");
  });

  it("should show admin nav items only for admin users", () => {
    renderApp(<AppLayout />, { auth: { user: { id: 1, email: "admin@local", displayName: "Admin", role: "admin", mustChangePassword: false } } });
    expect(screen.getByRole("link", { name: /users/i })).toBeInTheDocument();
  });

  it("should hide admin nav items for non-admin users", () => {
    renderApp(<AppLayout />, { auth: { user: { id: 2, email: "viewer@local", displayName: "Viewer", role: "viewer", mustChangePassword: false } } });
    expect(screen.queryByRole("link", { name: /users/i })).not.toBeInTheDocument();
  });

  it("should show theme selector with auto/light/dark options", () => {
    renderApp(<AppLayout />);
    expect(screen.getByRole("radiogroup", { name: /theme/i })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /auto/i })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /light/i })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /dark/i })).toBeInTheDocument();
  });

  it("should call logout and redirect on logout click", async () => {
    const user = userEvent.setup();
    const logoutFn = vi.fn();
    renderApp(<AppLayout />, { auth: { logout: logoutFn } });
    const buttons = screen.getAllByRole("button");
    const logoutBtn = buttons.find((btn) => btn.querySelector(".lucide-log-out") !== null);
    expect(logoutBtn).toBeDefined();
    await user.click(logoutBtn!);
    expect(logoutFn).toHaveBeenCalled();
  });

  it("should open mobile sidebar on menu button", async () => {
    const user = userEvent.setup();
    renderApp(<AppLayout />);
    const buttons = screen.getAllByRole("button");
    const menuButton = buttons.find((btn) => btn.querySelector(".lucide-menu") !== null);
    if (menuButton) {
      await user.click(menuButton);
    }
  });

  it("should have aria-label on sidebar aside", () => {
    renderApp(<AppLayout />);
    expect(screen.getByRole("complementary", { name: /sidebar/i })).toBeInTheDocument();
  });

  it("should have aria-label on main navigation", () => {
    renderApp(<AppLayout />);
    expect(screen.getByRole("navigation", { name: /main/i })).toBeInTheDocument();
  });

  it("should have aria-expanded on collapse toggle button", () => {
    const original = window.matchMedia;
    window.matchMedia = vi.fn().mockImplementation((query: string) => ({
      matches: query === "(min-width: 768px)",
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }));
    renderApp(<AppLayout />);
    const collapseBtn = screen.getByRole("button", { name: /collapse sidebar/i });
    expect(collapseBtn).toHaveAttribute("aria-expanded", "true");
    window.matchMedia = original;
  });

  it("should have theme radiogroup in sidebar", () => {
    renderApp(<AppLayout />);
    expect(screen.getByRole("radiogroup", { name: /theme/i })).toBeInTheDocument();
  });

  it("should have aria-label on logout button", () => {
    renderApp(<AppLayout />);
    expect(screen.getByRole("button", { name: /log out/i })).toBeInTheDocument();
  });

  it("should have main landmark with id for skip-to-content", () => {
    renderApp(<AppLayout />);
    expect(screen.getByRole("main")).toHaveAttribute("id", "main-content");
  });

  it("should have a skip-to-content link targeting main-content", () => {
    renderApp(<AppLayout />);
    const skipLink = screen.getByText(/skip to/i);
    expect(skipLink).toHaveAttribute("href", "#main-content");
    expect(skipLink.tagName).toBe("A");
  });

  it("should auto-collapse sidebar on tablet viewport", () => {
    const original = window.matchMedia;
    window.matchMedia = vi.fn().mockImplementation((query: string) => ({
      matches: query === "(min-width: 768px)" || query === "(min-width: 768px) and (max-width: 1023px)",
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }));
    renderApp(<AppLayout />);
    const aside = screen.getByRole("complementary", { name: /sidebar/i });
    expect(aside.className).toContain("w-[56px]");
    expect(screen.queryByRole("button", { name: /collapse sidebar/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /expand sidebar/i })).not.toBeInTheDocument();
    window.matchMedia = original;
  });

  it("should open keyboard shortcuts overlay when pressing ? and close via Escape", async () => {
    const user = userEvent.setup();
    renderApp(<AppLayout />);
    await user.keyboard("?");
    expect(screen.getByRole("dialog", { name: /keyboard shortcuts/i })).toBeInTheDocument();
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("dialog", { name: /keyboard shortcuts/i })).not.toBeInTheDocument();
  });

  it("should toggle keyboard shortcuts overlay off on second ? press", async () => {
    const user = userEvent.setup();
    renderApp(<AppLayout />);
    await user.keyboard("?");
    expect(screen.getByRole("dialog", { name: /keyboard shortcuts/i })).toBeInTheDocument();
    await user.keyboard("?");
    expect(screen.queryByRole("dialog", { name: /keyboard shortcuts/i })).not.toBeInTheDocument();
  });

  it("should open search when clicking the expanded search button", async () => {
    const user = userEvent.setup();
    renderApp(<AppLayout />);
    const searchButton = screen.getByText("Search").closest("button")!;
    await user.click(searchButton);
    expect(screen.getByRole("dialog", { name: /search/i })).toBeInTheDocument();
  });

  it("should show collapsed search button when sidebar is collapsed via localStorage", async () => {
    localStorage.setItem("gavel-sidebar-collapsed", "true");
    const original = window.matchMedia;
    window.matchMedia = vi.fn().mockImplementation((query: string) => ({
      matches: query === "(min-width: 768px)",
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }));
    const user = userEvent.setup();
    renderApp(<AppLayout />);
    const aside = screen.getByRole("complementary", { name: /sidebar/i });
    expect(aside.className).toContain("w-[56px]");
    const searchButton = screen.getByTitle("Search (⌘K)");
    expect(searchButton).toBeInTheDocument();
    await user.click(searchButton);
    expect(screen.getByRole("dialog", { name: /search/i })).toBeInTheDocument();
    localStorage.removeItem("gavel-sidebar-collapsed");
    window.matchMedia = original;
  });
});
