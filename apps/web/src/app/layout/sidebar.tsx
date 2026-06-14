import { useState, useEffect } from "react";
import { Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "@/entities/user/use-auth";
import { useTheme } from "@/shared/lib/use-theme";
import { useMediaQuery } from "@/shared/lib/use-media-query";
import { Button } from "@/shared/ui/button";
import { Avatar } from "@/shared/ui/avatar";
import { CommandPalette } from "@/shared/ui/command-palette";
import { KeyboardShortcutsOverlay } from "@/shared/ui/keyboard-shortcuts";
import { ThemeSelector } from "@/shared/ui/theme-selector";
import { cn } from "@/shared/lib/utils";
import { NavItem } from "./nav-item";
import { SectionLabel } from "./section-label";
import { useCollapsed } from "./use-collapsed";
import {
  Building2,
  Key,
  Users,
  LogOut,
  Menu,
  X,
  Sun,
  Moon,
  Search,
  PanelLeftClose,
  PanelLeftOpen,
} from "lucide-react";

export function AppLayout() {
  const { user, logout } = useAuth();
  const { resolved, toggleTheme } = useTheme();
  const navigate = useNavigate();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [searchOpen, setSearchOpen] = useState(false);
  const [shortcutsOpen, setShortcutsOpen] = useState(false);
  const { collapsed: userCollapsed, toggle: toggleCollapsed } = useCollapsed();
  const isDesktop = useMediaQuery("(min-width: 768px)");
  const isTablet = useMediaQuery("(min-width: 768px) and (max-width: 1023px)");
  const collapsed = isTablet || userCollapsed;

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setSearchOpen(true);
        return;
      }
      const target = e.target as HTMLElement;
      if (target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.isContentEditable) return;
      if (e.key === "?" && !e.metaKey && !e.ctrlKey) {
        e.preventDefault();
        setShortcutsOpen((prev) => !prev);
      }
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, []);

  const handleLogout = async () => {
    await logout();
    navigate("/login");
  };

  const closeSidebar = () => setSidebarOpen(false);

  return (
    <div className="flex h-screen">
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[100] focus:rounded-md focus:bg-primary focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:text-primary-foreground focus:shadow-floating"
      >
        Skip to main content
      </a>
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 md:hidden"
          onClick={closeSidebar}
        />
      )}

      <aside
        aria-label="Sidebar"
        className={cn(
          "fixed inset-y-0 left-0 z-50 flex flex-col border-r border-border bg-card transition-all duration-normal ease-out-expo md:static md:translate-x-0",
          collapsed ? "w-[56px]" : "w-[220px]",
          sidebarOpen ? "translate-x-0" : "-translate-x-full",
        )}
      >
        <div className={cn(
          "flex h-[52px] items-center border-b border-border",
          collapsed ? "justify-center px-2" : "justify-between px-3",
        )}>
          {!collapsed && (
            <div className="flex items-center gap-2">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" className="text-foreground">
                <path d="M4 14 L10 8 L14 12 L20 6" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                <circle cx="20" cy="6" r="1.8" className="fill-primary"/>
              </svg>
              <span className="text-sm font-semibold tracking-tight">Gavel</span>
            </div>
          )}
          {collapsed && !isTablet && (
            <button
              onClick={toggleCollapsed}
              className="rounded-lg p-1 text-muted-foreground/60 hover:bg-muted/50 hover:text-foreground transition-colors duration-fast"
              aria-label="Expand sidebar"
              aria-expanded={false}
            >
              <PanelLeftOpen className="h-4 w-4" />
            </button>
          )}
          {collapsed && isTablet && (
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" className="text-foreground">
              <path d="M4 14 L10 8 L14 12 L20 6" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              <circle cx="20" cy="6" r="1.8" className="fill-primary"/>
            </svg>
          )}
          {!collapsed && (
            <div className="flex items-center gap-1">
              {isDesktop && !isTablet && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7"
                  onClick={toggleCollapsed}
                  aria-label="Collapse sidebar"
                  aria-expanded={true}
                >
                  <PanelLeftClose className="h-3.5 w-3.5" />
                </Button>
              )}
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7 md:hidden"
                onClick={closeSidebar}
                aria-label="Close sidebar"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          )}
        </div>

        <nav
          aria-label="Main navigation"
          className={cn(
            "flex-1 space-y-0.5 overflow-y-auto pt-3",
            collapsed ? "px-2" : "px-3",
          )}
          onClick={closeSidebar}
        >
          {!collapsed && (
            <button
              onClick={() => setSearchOpen(true)}
              className="mb-2.5 flex w-full items-center gap-2 rounded-lg border border-border bg-muted px-2.5 py-[7px] text-xs text-muted-foreground hover:bg-muted/80 transition-colors duration-fast"
            >
              <Search className="h-3 w-3" />
              <span className="flex-1 text-left">Search</span>
              <span className="font-mono text-2xs text-muted-foreground/50">⌘K</span>
            </button>
          )}
          {collapsed && (
            <button
              onClick={() => setSearchOpen(true)}
              title="Search (⌘K)"
              className="mb-2.5 flex w-full items-center justify-center rounded-lg border border-border bg-muted py-[7px] text-xs text-muted-foreground hover:bg-muted/80 transition-colors duration-fast"
            >
              <Search className="h-3 w-3" />
            </button>
          )}

          <NavItem to="/gavelspaces" icon={<Building2 className="h-3.5 w-3.5" />} label="Gavelspaces" collapsed={collapsed} onClick={closeSidebar} />

          <div className="my-2 h-px bg-border" />
          <SectionLabel collapsed={collapsed}>Settings</SectionLabel>
          <NavItem to="/tokens" icon={<Key className="h-3.5 w-3.5" />} label="API Tokens" collapsed={collapsed} onClick={closeSidebar} />
          {user?.role === "admin" && (
            <NavItem to="/admin/users" icon={<Users className="h-3.5 w-3.5" />} label="Users" collapsed={collapsed} onClick={closeSidebar} />
          )}
        </nav>

        <div className={cn("border-t border-border", collapsed ? "px-2 py-2" : "px-3 py-2")}>
          {collapsed ? (
            <button
              onClick={toggleTheme}
              className="flex w-full items-center justify-center rounded-lg py-1.5 text-muted-foreground hover:bg-muted/50 hover:text-foreground transition-colors duration-fast"
              aria-label="Toggle theme"
              title="Toggle theme"
            >
              {resolved === "dark" ? <Sun className="h-3.5 w-3.5" /> : <Moon className="h-3.5 w-3.5" />}
            </button>
          ) : (
            <ThemeSelector />
          )}
        </div>

        <div className={cn("border-t border-border", collapsed ? "p-2" : "p-3")}>
          {collapsed ? (
            <div className="flex justify-center" title={user?.displayName ?? user?.email}>
              <Avatar initials={user?.displayName?.slice(0, 2).toUpperCase() ?? "U"} tone="violet" />
            </div>
          ) : (
            <div className="flex items-center gap-2.5 rounded-lg border border-border bg-muted p-2.5">
              <Avatar initials={user?.displayName?.slice(0, 2).toUpperCase() ?? "U"} tone="violet" />
              <div className="min-w-0 flex-1">
                <div className="truncate text-xs font-medium">{user?.displayName ?? user?.email}</div>
                <div className="text-2xs text-muted-foreground">{user?.role}</div>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 shrink-0"
                onClick={handleLogout}
                aria-label="Log out"
              >
                <LogOut className="h-3 w-3" />
              </Button>
            </div>
          )}
        </div>
      </aside>

      <div className="flex flex-1 flex-col overflow-hidden">
        <header className="flex h-14 items-center border-b border-border px-4 md:hidden">
          <Button variant="ghost" size="icon" onClick={() => setSidebarOpen(true)} aria-label="Open sidebar">
            <Menu className="h-5 w-5" />
          </Button>
          <span className="ml-3 text-lg font-bold tracking-tight">Gavel</span>
        </header>
        <main id="main-content" className="flex-1 overflow-auto bg-background">
          <Outlet />
        </main>
      </div>

      <CommandPalette open={searchOpen} onClose={() => setSearchOpen(false)} />
      <KeyboardShortcutsOverlay open={shortcutsOpen} onClose={() => setShortcutsOpen(false)} />
    </div>
  );
}
