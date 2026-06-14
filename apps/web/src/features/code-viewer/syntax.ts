import type Prism from "prismjs";

const EXTENSION_TO_LANGUAGE: Record<string, string> = {
  go: "go",
  ts: "typescript",
  tsx: "tsx",
  js: "javascript",
  jsx: "jsx",
  rs: "rust",
  py: "python",
  java: "java",
  kt: "kotlin",
};

export function inferLanguageFromExtension(filePath: string): string {
  const match = /\.([^.\\/]+)$/.exec(filePath);
  if (!match) return "";
  return EXTENSION_TO_LANGUAGE[match[1].toLowerCase()] ?? "";
}

type PrismModule = typeof Prism;

let prismPromise: Promise<PrismModule> | null = null;
const loadedLanguages = new Set<string>();

async function loadPrism(): Promise<PrismModule> {
  if (!prismPromise) {
    prismPromise = import("prismjs").then((mod) => (mod as { default?: PrismModule }).default ?? (mod as unknown as PrismModule));
  }
  return prismPromise;
}

async function loadLanguageGrammar(language: string): Promise<boolean> {
  if (loadedLanguages.has(language)) return true;
  try {
    switch (language) {
      case "go":
        await import("prismjs/components/prism-go");
        break;
      case "typescript":
        await import("prismjs/components/prism-typescript");
        break;
      case "tsx":
        await import("prismjs/components/prism-typescript");
        await import("prismjs/components/prism-jsx");
        await import("prismjs/components/prism-tsx");
        break;
      case "javascript":
        await import("prismjs/components/prism-javascript");
        break;
      case "jsx":
        await import("prismjs/components/prism-jsx");
        break;
      case "rust":
        await import("prismjs/components/prism-rust");
        break;
      case "python":
        await import("prismjs/components/prism-python");
        break;
      case "java":
        await import("prismjs/components/prism-java");
        break;
      case "kotlin":
        await import("prismjs/components/prism-kotlin");
        break;
      default:
        return false;
    }
  } catch {
    return false;
  }
  loadedLanguages.add(language);
  return true;
}

export async function highlightSource(
  source: string,
  language: string,
): Promise<string[] | null> {
  if (!language) return null;
  const prism = await loadPrism();
  const ok = await loadLanguageGrammar(language);
  if (!ok) return null;
  const grammar = prism.languages[language];
  if (!grammar) return null;
  const html = prism.highlight(source, grammar, language);
  return html.split("\n");
}
