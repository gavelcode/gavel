import { useEffect, useState } from "react";
import { highlightSource } from "./syntax";

interface CachedHighlight {
  source: string;
  language: string;
  lines: string[];
}

export function useAsyncSourceHighlights(
  source: string,
  language: string | undefined,
): string[] | null {
  const [cache, setCache] = useState<CachedHighlight | null>(null);

  useEffect(() => {
    if (!language || source.length === 0) return;
    let cancelled = false;
    highlightSource(source, language)
      .then((lines) => {
        if (!cancelled && lines) {
          setCache({ source, language, lines });
        }
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [source, language]);

  if (cache && cache.source === source && cache.language === language) {
    return cache.lines;
  }
  return null;
}
