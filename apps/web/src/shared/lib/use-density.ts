import { createContext, useContext } from "react";

export type Density = "comfortable" | "compact" | "dense";

export interface DensityContextValue {
  density: Density;
  setDensity: (d: Density) => void;
}

export const DENSITY_STORAGE_KEY = "gavel-density";

export const DensityContext = createContext<DensityContextValue>({
  density: "compact",
  setDensity: () => {},
});

export function useDensity() {
  return useContext(DensityContext);
}
