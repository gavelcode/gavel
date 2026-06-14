import { createContext } from "react";

type Theme = "light" | "dark" | "system";

export interface ThemeContextValue {
  theme: Theme;
  resolved: "light" | "dark";
  highContrast: boolean;
  setTheme: (t: Theme) => void;
  toggleTheme: () => void;
  setHighContrast: (v: boolean) => void;
}

export const ThemeContext = createContext<ThemeContextValue>({
  theme: "system",
  resolved: "light",
  highContrast: false,
  setTheme: () => {},
  toggleTheme: () => {},
  setHighContrast: () => {},
});
