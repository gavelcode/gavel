import { useCallback, useEffect, useState, type ReactNode } from "react";

import {
  DENSITY_STORAGE_KEY,
  DensityContext,
  type Density,
} from "./use-density";

export function DensityProvider({ children }: { children: ReactNode }) {
  const [density, setDensityState] = useState<Density>(() => {
    if (typeof window === "undefined") return "compact";
    return (localStorage.getItem(DENSITY_STORAGE_KEY) as Density) ?? "compact";
  });

  useEffect(() => {
    document.documentElement.dataset.density = density;
  }, [density]);

  const setDensity = useCallback((d: Density) => {
    setDensityState(d);
    localStorage.setItem(DENSITY_STORAGE_KEY, d);
  }, []);

  return (
    <DensityContext.Provider value={{ density, setDensity }}>
      {children}
    </DensityContext.Provider>
  );
}
