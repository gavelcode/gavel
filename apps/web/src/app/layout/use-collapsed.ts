import { useState } from "react";

const STORAGE_KEY = "gavel-sidebar-collapsed";

export function useCollapsed() {
  const [collapsed, setCollapsed] = useState(() => {
    try {
      return localStorage.getItem(STORAGE_KEY) === "true";
    } catch {
      return false;
    }
  });

  function toggle() {
    setCollapsed((prev) => {
      const next = !prev;
      try {
        localStorage.setItem(STORAGE_KEY, String(next));
      } catch { /* noop */ }
      return next;
    });
  }

  return { collapsed, toggle };
}
