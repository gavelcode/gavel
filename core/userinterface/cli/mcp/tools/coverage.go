package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
)

const (
	verdictPass    = "PASS"
	verdictFail    = "FAIL"
	statusNew      = "new"
	statusNewLabel = "NEW"
)

type CoverageInput struct {
	Gavelspace string   `json:"gavelspace,omitempty" jsonschema:"Absolute path to a gavelspace directory (omit to use the directory where Claude Code is running)"`
	Project    string   `json:"project,omitempty"    jsonschema:"Project name (analyzes all projects if omitted)"`
	Files      []string `json:"files,omitempty"      jsonschema:"Filter the per-file breakdown to these file paths (relative to workspace root). When set, uncovered line ranges are listed for each file."`
	Packages   []string `json:"packages,omitempty"   jsonschema:"Filter the per-file breakdown to whole package trees (relative to workspace root). A trailing '/...' matches the package and its subpackages; without it, only files directly in that package. Mutually exclusive with 'files'. Same output as 'files', with uncovered line ranges."`
	Diff       bool     `json:"diff,omitempty"       jsonschema:"Show per-file coverage change since the last green baseline (only files whose coverage moved, plus files with no baseline marked '(new)'). Reports 'No previous baseline' on the first run."`
}

func RegisterCoverage(server *mcp.Server, cli *executor.CLI) {
	mcp.AddTool(server, &mcp.Tool{
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
		Name:        "gavel_coverage",
		Description: "Get code coverage for a project. Runs a full analysis and reports the overall percentage plus a per-file breakdown. Pass 'files' to focus on specific files and see their uncovered line ranges.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CoverageInput) (*mcp.CallToolResult, any, error) {
		result, err := runCoverage(ctx, cli, input)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}}, nil, nil
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil, nil
	})
}

func runCoverage(ctx context.Context, cli *executor.CLI, input CoverageInput) (string, error) {
	if err := validateCoverageInput(input); err != nil {
		return "", err
	}

	args := []string{"judge", "--no-baseline-update"}
	if input.Project != "" {
		args = append(args, "--project", input.Project)
	}

	output, _, err := cli.RunInJSON(ctx, input.Gavelspace, args...)
	if err != nil {
		return "", fmt.Errorf("execute gavel judge: %w", err)
	}

	var resp judgeResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return "", fmt.Errorf("parse judge output: %w", err)
	}

	if input.Diff {
		return formatCoverageDiff(resp), nil
	}

	files := input.Files
	if len(input.Packages) > 0 {
		files, err = resolvePackages(resp, input.Packages)
		if err != nil {
			return "", err
		}
	}

	return formatCoverage(resp, files), nil
}

func formatCoverageDiff(resp judgeResponse) string {
	var builder strings.Builder
	for _, proj := range resp.Projects {
		fmt.Fprintf(&builder, "## %s\n", proj.Name)
		if proj.CoveragePercent != nil {
			fmt.Fprintf(&builder, "Coverage: %.1f%%\n", *proj.CoveragePercent)
		}
		if proj.Delta == nil || !proj.Delta.HasPrevious {
			builder.WriteString("No previous baseline to diff against.\n\n")
			continue
		}
		writeCoverageDiff(&builder, proj.CoverageByFile)
		builder.WriteString("\n")
	}
	return builder.String()
}

func writeCoverageDiff(builder *strings.Builder, byFile []judgeFileCoverage) {
	var changed bool
	for _, file := range byFile {
		switch {
		case file.IsNew:
			fmt.Fprintf(builder, "  %s — %.1f%% (new)\n", file.FilePath, file.Percent)
			changed = true
		case file.CoverageDelta != nil && *file.CoverageDelta != 0 && file.PreviousPercent != nil:
			fmt.Fprintf(builder, "  %s — %.1f%% (%+.1f%% from baseline %.1f%%)\n", file.FilePath, file.Percent, *file.CoverageDelta, *file.PreviousPercent)
			changed = true
		}
	}
	if !changed {
		builder.WriteString("No coverage changes since baseline.\n")
	}
}

func validateCoverageInput(input CoverageInput) error {
	if len(input.Packages) > 0 && len(input.Files) > 0 {
		return fmt.Errorf("'packages' and 'files' are mutually exclusive; pass only one")
	}
	return nil
}

func resolvePackages(resp judgeResponse, packages []string) ([]string, error) {
	var coveredPaths []string
	for _, proj := range resp.Projects {
		for _, fc := range proj.CoverageByFile {
			coveredPaths = append(coveredPaths, fc.FilePath)
		}
	}

	matched := make(map[string]bool)
	for _, pkg := range packages {
		var found bool
		for _, filePath := range coveredPaths {
			if packageMatches(pkg, filePath) {
				matched[filePath] = true
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("no coverage data for package %q (typo, no production code, or outside the analyzed project?)", pkg)
		}
	}

	files := make([]string, 0, len(matched))
	for filePath := range matched {
		files = append(files, filePath)
	}
	sort.Strings(files)
	return files, nil
}

func packageMatches(pkg, filePath string) bool {
	if strings.HasSuffix(pkg, "/...") {
		base := strings.TrimSuffix(pkg, "...")
		return strings.HasPrefix(filePath, base)
	}
	return path.Dir(filePath) == pkg
}

func formatCoverage(resp judgeResponse, files []string) string {
	var builder strings.Builder
	for _, proj := range resp.Projects {
		fmt.Fprintf(&builder, "## %s\n", proj.Name)
		if proj.CoveragePercent != nil {
			fmt.Fprintf(&builder, "Coverage: %.1f%%\n", *proj.CoveragePercent)
		} else {
			builder.WriteString("Coverage: not collected (use full analysis, not --quick)\n")
		}
		for _, ruling := range proj.Rulings {
			if strings.Contains(ruling.Subtype, "coverage") {
				status := verdictPass
				if !ruling.Passed {
					status = verdictFail
				}
				fmt.Fprintf(&builder, "Rule %s: %s — %s\n", ruling.Subtype, status, ruling.Detail)
			}
		}
		if len(files) > 0 {
			writeFileBreakdown(&builder, proj.CoverageByFile, files)
		} else if proj.CoverageTree != nil {
			writeCoverageTree(&builder, proj.CoverageTree, 0)
		} else {
			writeFileBreakdown(&builder, proj.CoverageByFile, nil)
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

func writeCoverageTree(builder *strings.Builder, node *judgeCoverageNode, depth int) {
	indent := strings.Repeat("  ", depth)
	for _, child := range node.Children {
		fmt.Fprintf(builder, "%s%s/ — %.1f%% (%d/%d)\n", indent, child.Path, child.Percent, child.CoveredLines, child.TotalLines)
		for _, f := range child.Files {
			fmt.Fprintf(builder, "%s  %s — %.1f%% (%d/%d)\n", indent, f.Name, f.Percent, f.CoveredLines, f.TotalLines)
		}
		writeCoverageTree(builder, &child, depth+1)
	}
}

func writeFindingsTree(builder *strings.Builder, node *judgeFindingsNode, depth int) {
	indent := strings.Repeat("  ", depth)
	for _, child := range node.Children {
		fmt.Fprintf(builder, "%s%s/ — %d findings", indent, child.Path, child.Count)
		if len(child.BySeverity) > 0 {
			fmt.Fprintf(builder, " (%s)", formatSeverities(child.BySeverity))
		}
		builder.WriteString("\n")
		for _, f := range child.Files {
			fmt.Fprintf(builder, "%s  %s — %d", indent, f.Name, f.Count)
			if len(f.BySeverity) > 0 {
				fmt.Fprintf(builder, " (%s)", formatSeverities(f.BySeverity))
			}
			builder.WriteString("\n")
		}
		writeFindingsTree(builder, &child, depth+1)
	}
}

func formatSeverities(m map[string]int) string {
	var parts []string
	for _, sev := range []string{"error", "warning", "note"} {
		if n, ok := m[sev]; ok {
			parts = append(parts, fmt.Sprintf("%d %s", n, sev))
		}
	}
	return strings.Join(parts, ", ")
}

func writeFileBreakdown(builder *strings.Builder, byFile []judgeFileCoverage, filter []string) {
	wanted := fileFilter(filter)
	var rows []judgeFileCoverage
	for _, fc := range byFile {
		if wanted != nil && !wanted[fc.FilePath] {
			continue
		}
		rows = append(rows, fc)
	}
	if len(rows) == 0 {
		return
	}

	builder.WriteString("\nBy file:\n")
	for _, fc := range rows {
		fmt.Fprintf(builder, "  %s — %.1f%% (%d/%d)", fc.FilePath, fc.Percent, fc.CoveredLines, fc.TotalLines)
		if wanted != nil && len(fc.Uncovered) > 0 {
			builder.WriteString(", uncovered: ")
			builder.WriteString(compressRanges(fc.Uncovered))
		}
		builder.WriteString("\n")
	}
}

func fileFilter(filter []string) map[string]bool {
	if len(filter) == 0 {
		return nil
	}
	set := make(map[string]bool, len(filter))
	for _, f := range filter {
		set[f] = true
	}
	return set
}

func compressRanges(lines []int) string {
	if len(lines) == 0 {
		return ""
	}
	var parts []string
	start, prev := lines[0], lines[0]
	flush := func() {
		if start == prev {
			parts = append(parts, strconv.Itoa(start))
			return
		}
		parts = append(parts, fmt.Sprintf("%d-%d", start, prev))
	}
	for _, line := range lines[1:] {
		if line == prev+1 {
			prev = line
			continue
		}
		flush()
		start, prev = line, line
	}
	flush()
	return strings.Join(parts, ", ")
}
