import {
  useCallback,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { ThemeContext } from "./theme-context";

type Theme = "light" | "dark" | "system";

function resolveTheme(theme: Theme): "light" | "dark" {
  if (theme !== "system") return theme;
  if (typeof window === "undefined") return "light";
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

const STORAGE_KEY = "gavel-theme";
const HC_STORAGE_KEY = "gavel-high-contrast";

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<Theme>(() => {
    if (typeof window === "undefined") return "system";
    return (localStorage.getItem(STORAGE_KEY) as Theme) ?? "system";
  });

  const [highContrast, setHighContrastState] = useState(() => {
    if (typeof window === "undefined") return false;
    return localStorage.getItem(HC_STORAGE_KEY) === "true";
  });

  const resolved = resolveTheme(theme);

  useEffect(() => {
    const root = document.documentElement;
    root.classList.toggle("dark", resolved === "dark");
    root.classList.toggle("high-contrast", highContrast);
  }, [resolved, highContrast]);

  useEffect(() => {
    if (theme === "system") return;
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const handler = () => setThemeState((t) => (t === "system" ? t : t));
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, [theme]);

  const setTheme = useCallback((t: Theme) => {
    setThemeState(t);
    localStorage.setItem(STORAGE_KEY, t);
  }, []);

  const toggleTheme = useCallback(() => {
    setThemeState((prev) => {
      const next = resolveTheme(prev) === "dark" ? "light" : "dark";
      localStorage.setItem(STORAGE_KEY, next);
      return next;
    });
  }, []);

  const setHighContrast = useCallback((v: boolean) => {
    setHighContrastState(v);
    localStorage.setItem(HC_STORAGE_KEY, String(v));
  }, []);

  return (
    <ThemeContext.Provider value={{ theme, resolved, highContrast, setTheme, toggleTheme, setHighContrast }}>
      {children}
    </ThemeContext.Provider>
  );
}
