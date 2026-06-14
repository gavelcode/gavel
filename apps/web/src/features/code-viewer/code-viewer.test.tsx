import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CodeViewer } from "./code-viewer";
import type { Finding } from "@/entities/finding/model";

function makeFinding(overrides: Partial<Finding> = {}): Finding {
  return {
    tool: "golangci-lint",
    ruleId: "unused",
    severity: "warning",
    filePath: "main.go",
    line: 1,
    message: "unused variable",
    fingerprint: "fp-1",
    status: "new",
    source: "linter",
    commitSha: "abc123",
    projectKey: "test-project",
    ...overrides,
  };
}

describe("CodeViewer", () => {
  it("renders one row per line, numbered from 1, with matching text content", () => {
    render(<CodeViewer source={"alpha\nbeta\ngamma"} />);

    const lines = screen.getAllByTestId("code-line");
    expect(lines).toHaveLength(3);

    expect(lines[0]).toHaveAttribute("data-line-number", "1");
    expect(lines[1]).toHaveAttribute("data-line-number", "2");
    expect(lines[2]).toHaveAttribute("data-line-number", "3");

    expect(lines[0]).toHaveTextContent("alpha");
    expect(lines[1]).toHaveTextContent("beta");
    expect(lines[2]).toHaveTextContent("gamma");
  });

  it("preserves leading indentation and whitespace in line text", () => {
    const source = "func main() {\n\tfmt.Println(\"hi\")\n    return\n}";
    render(<CodeViewer source={source} />);

    const lines = screen.getAllByTestId("code-line");
    expect(lines).toHaveLength(4);

    const textCell = (row: HTMLElement) =>
      row.querySelector("[data-testid='code-line-text']") as HTMLElement;

    expect(textCell(lines[1]).textContent).toBe("\tfmt.Println(\"hi\")");
    expect(textCell(lines[2]).textContent).toBe("    return");

    const styles = window.getComputedStyle(textCell(lines[0]));
    expect(styles.whiteSpace).toBe("pre");
  });

  it("renders an empty-file placeholder when source is empty", () => {
    render(<CodeViewer source="" />);

    expect(screen.queryAllByTestId("code-line")).toHaveLength(0);
    expect(screen.getByTestId("code-viewer-empty")).toHaveTextContent(/empty file/i);
  });

  it("marks the active line with data-active when activeLine is set", () => {
    render(<CodeViewer source={"a\nb\nc\nd\ne\nf"} activeLine={5} />);

    const lines = screen.getAllByTestId("code-line");
    expect(lines[4]).toHaveAttribute("data-active", "true");

    for (const idx of [0, 1, 2, 3, 5]) {
      expect(lines[idx]).not.toHaveAttribute("data-active", "true");
    }
  });

  describe("coverage gutter", () => {
    function gutterOf(row: HTMLElement) {
      return row.querySelector("[data-testid='coverage-gutter']") as HTMLElement;
    }

    it("covered line gets a success class on its gutter cell", () => {
      const coverage = new Map<number, "covered" | "uncovered" | "none">([
        [1, "covered"],
      ]);
      render(<CodeViewer source={"only one line"} coverage={coverage} />);

      const cell = gutterOf(screen.getAllByTestId("code-line")[0]);
      expect(cell.className).toMatch(/success/);
      expect(cell).toHaveAttribute("data-coverage", "covered");
      expect(cell).toHaveAttribute("aria-label", "Coverage: covered");
    });

    it("uncovered line gets a danger class on its gutter cell", () => {
      const coverage = new Map<number, "covered" | "uncovered" | "none">([
        [1, "uncovered"],
      ]);
      render(<CodeViewer source={"only one line"} coverage={coverage} />);

      const cell = gutterOf(screen.getAllByTestId("code-line")[0]);
      expect(cell.className).toMatch(/danger/);
      expect(cell).toHaveAttribute("data-coverage", "uncovered");
      expect(cell).toHaveAttribute("aria-label", "Coverage: uncovered");
    });

    it("line not present in the coverage map has no color class on its gutter cell", () => {
      render(<CodeViewer source={"only one line"} coverage={new Map()} />);

      const cell = gutterOf(screen.getAllByTestId("code-line")[0]);
      expect(cell.className).not.toMatch(/success/);
      expect(cell.className).not.toMatch(/danger/);
      expect(cell).toHaveAttribute("data-coverage", "none");
    });

    it("renders the right class per line in a mixed file", () => {
      const coverage = new Map<number, "covered" | "uncovered" | "none">([
        [1, "covered"],
        [2, "uncovered"],
        [3, "none"],
        [4, "covered"],
        [5, "uncovered"],
      ]);
      render(
        <CodeViewer source={"a\nb\nc\nd\ne"} coverage={coverage} />,
      );

      const lines = screen.getAllByTestId("code-line");
      const states = lines.map((row) => gutterOf(row).getAttribute("data-coverage"));
      expect(states).toEqual([
        "covered",
        "uncovered",
        "none",
        "covered",
        "uncovered",
      ]);
      expect(gutterOf(lines[0]).className).toMatch(/success/);
      expect(gutterOf(lines[1]).className).toMatch(/danger/);
      expect(gutterOf(lines[3]).className).toMatch(/success/);
      expect(gutterOf(lines[4]).className).toMatch(/danger/);
    });
  });

  describe("finding markers", () => {
    function markersOf(row: HTMLElement) {
      return Array.from(
        row.querySelectorAll("[data-testid='finding-marker']"),
      ) as HTMLElement[];
    }

    it("renders a marker on the row matching the finding line", () => {
      const finding = makeFinding({ line: 12, ruleId: "G101" });
      const source = Array.from({ length: 20 }, (_, i) => `line ${i + 1}`).join(
        "\n",
      );
      render(<CodeViewer source={source} findings={[finding]} />);

      const all = screen.getAllByTestId("finding-marker");
      expect(all).toHaveLength(1);

      const lines = screen.getAllByTestId("code-line");
      expect(markersOf(lines[11])).toHaveLength(1);
      expect(markersOf(lines[0])).toHaveLength(0);

      const marker = all[0];
      const label = marker.getAttribute("aria-label") ?? "";
      expect(label).toMatch(/warning/i);
      expect(label).toContain("G101");
    });

    it("renders one marker per finding when multiple findings target the same file", () => {
      const findings = [
        makeFinding({ line: 3, fingerprint: "a" }),
        makeFinding({ line: 7, fingerprint: "b" }),
        makeFinding({ line: 12, fingerprint: "c" }),
      ];
      const source = Array.from({ length: 15 }, (_, i) => `line ${i + 1}`).join(
        "\n",
      );
      render(<CodeViewer source={source} findings={findings} />);

      expect(screen.getAllByTestId("finding-marker")).toHaveLength(3);

      const lines = screen.getAllByTestId("code-line");
      expect(markersOf(lines[2])).toHaveLength(1);
      expect(markersOf(lines[6])).toHaveLength(1);
      expect(markersOf(lines[11])).toHaveLength(1);
    });

    it("uses distinct color classes per severity", () => {
      const findings = [
        makeFinding({ line: 1, severity: "error", fingerprint: "e" }),
        makeFinding({ line: 2, severity: "warning", fingerprint: "w" }),
        makeFinding({ line: 3, severity: "note", fingerprint: "n" }),
      ];
      render(
        <CodeViewer source={"l1\nl2\nl3"} findings={findings} />,
      );

      const [errMarker, warnMarker, noteMarker] =
        screen.getAllByTestId("finding-marker");
      expect(errMarker.className).toMatch(/danger/);
      expect(warnMarker.className).toMatch(/warning/);
      expect(noteMarker.className).toMatch(/muted/);
    });

    it("invokes onFindingClick with the clicked finding", async () => {
      const finding = makeFinding({ line: 2, fingerprint: "pick-me" });
      const onFindingClick = vi.fn();
      const user = userEvent.setup();

      render(
        <CodeViewer
          source={"a\nb\nc"}
          findings={[finding]}
          onFindingClick={onFindingClick}
        />,
      );

      await user.click(screen.getByTestId("finding-marker"));
      expect(onFindingClick).toHaveBeenCalledTimes(1);
      expect(onFindingClick).toHaveBeenCalledWith(finding);
    });
  });

  describe("syntax highlighting", () => {
    it("renders plain text initially and applies token spans for a supported language", async () => {
      const source = 'package main\n\nfunc main() {\n  println("hi")\n}\n';
      const { findByText, container } = render(
        <CodeViewer source={source} language="go" />,
      );

      await findByText("main", { selector: ".token" }, { timeout: 3000 });
      expect(container.querySelectorAll(".token").length).toBeGreaterThan(0);
    });

    it("renders plain text without crashing when the language is unsupported", () => {
      const { container } = render(
        <CodeViewer source={"DISPLAY 'hi'."} language="cobol" />,
      );

      expect(container.querySelectorAll(".token").length).toBe(0);
      expect(container.querySelector("[data-testid='code-line-text']")?.textContent).toBe(
        "DISPLAY 'hi'.",
      );
    });
  });
});
