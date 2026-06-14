import { useCallback, useEffect, useRef } from "react";

interface ListKeyboardNavOptions {
  itemCount: number;
  activeIndex: number;
  onActiveChange: (index: number) => void;
  onSelect?: (index: number) => void;
}

export function useListKeyboardNav({
  itemCount,
  activeIndex,
  onActiveChange,
  onSelect,
}: ListKeyboardNavOptions) {
  const containerRef = useRef<HTMLDivElement>(null);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      if (target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.isContentEditable) return;

      if (e.key === "j" || e.key === "ArrowDown") {
        e.preventDefault();
        const next = Math.min(activeIndex + 1, itemCount - 1);
        onActiveChange(next);
      } else if (e.key === "k" || e.key === "ArrowUp") {
        e.preventDefault();
        if (activeIndex > 0) onActiveChange(activeIndex - 1);
      } else if (e.key === "Enter" && activeIndex >= 0) {
        e.preventDefault();
        onSelect?.(activeIndex);
      }
    },
    [activeIndex, itemCount, onActiveChange, onSelect],
  );

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    el.addEventListener("keydown", handleKeyDown);
    return () => el.removeEventListener("keydown", handleKeyDown);
  }, [handleKeyDown]);

  return { containerRef };
}
