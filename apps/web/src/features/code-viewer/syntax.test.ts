import { highlightSource, inferLanguageFromExtension } from "./syntax";

describe("inferLanguageFromExtension", () => {
  it.each([
    ["main.go", "go"],
    ["src/index.ts", "typescript"],
    ["src/app.tsx", "tsx"],
    ["lib.rs", "rust"],
    ["script.py", "python"],
    ["Foo.java", "java"],
    ["Bar.kt", "kotlin"],
  ])("maps %s -> %s", (path, expected) => {
    expect(inferLanguageFromExtension(path)).toBe(expected);
  });

  it("returns empty string for unknown extensions", () => {
    expect(inferLanguageFromExtension("notes.cobol")).toBe("");
    expect(inferLanguageFromExtension("README")).toBe("");
    expect(inferLanguageFromExtension("")).toBe("");
  });
});

describe("highlightSource", () => {
  it("returns null when language is empty", async () => {
    const result = await highlightSource("anything", "");
    expect(result).toBeNull();
  });

  it("returns null when the language is unsupported (falls back to plain text in caller)", async () => {
    const result = await highlightSource('IDENTIFICATION DIVISION.\n', "cobol");
    expect(result).toBeNull();
  });

  it("returns per-line HTML with token spans when tokenizing Go", async () => {
    const source = 'package main\n\nfunc main() {\n  println("hi")\n}\n';
    const lines = await highlightSource(source, "go");

    expect(lines).not.toBeNull();
    expect(lines!.length).toBeGreaterThanOrEqual(5);
    const joined = lines!.join("\n");
    expect(joined).toMatch(/class="token/);
    expect(joined).toMatch(/keyword/);
  });

  it.each([
    ["typescript", 'const x: number = 1;\n'],
    ["tsx", 'const App = () => <div />;\n'],
    ["javascript", 'const x = 1;\n'],
    ["jsx", 'const App = () => <div />;\n'],
    ["rust", 'fn main() {}\n'],
    ["python", 'def hello():\n  pass\n'],
    ["java", 'class Main {}\n'],
    ["kotlin", 'fun main() {}\n'],
  ])("highlights %s source", async (language, source) => {
    const lines = await highlightSource(source, language);
    expect(lines).not.toBeNull();
    expect(lines!.length).toBeGreaterThanOrEqual(1);
  });
});
