package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
)

type JudgeInput struct {
	Gavelspace string `json:"gavelspace,omitempty" jsonschema:"Absolute path to a gavelspace directory (omit to use the directory where Claude Code is running)"`
	Project    string `json:"project,omitempty"    jsonschema:"Analyze only this project (omit for all projects)"`
	Quick      bool   `json:"quick,omitempty"      jsonschema:"Skip coverage and architecture checks (findings only)"`
	Affected   bool   `json:"affected,omitempty"   jsonschema:"Analyze only targets affected by changed files. Implies --quick."`
}

func RegisterJudge(server *mcp.Server, cli *executor.CLI) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gavel_judge",
		Description: "Run static analyzers and evaluate the quality gate for configured projects. Returns verdict (pass/fail), findings count, coverage, and rule-by-rule results.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input JudgeInput) (*mcp.CallToolResult, any, error) {
		result, err := runJudge(ctx, cli, input)
		if err != nil {
			errResult := &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}}
			return errResult, nil, nil
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil, nil
	})
}

func runJudge(ctx context.Context, cli *executor.CLI, input JudgeInput) (string, error) {
	args := []string{"judge"}
	if input.Project != "" {
		args = append(args, "--project", input.Project)
	}
	if input.Affected {
		args = append(args, "--affected")
	}
	if input.Quick {
		args = append(args, "--quick")
	}

	output, exitCode, err := cli.RunInJSON(ctx, input.Gavelspace, args...)
	if err != nil {
		return "", fmt.Errorf("execute gavel judge: %w", err)
	}

	return formatJudgeOutput(output, exitCode)
}

type judgeResponse struct {
	Projects []judgeProject `json:"projects"`
}

type judgeProject struct {
	Name            string              `json:"name"`
	Verdict         string              `json:"verdict"`
	FindingsCount   int                 `json:"findings_count"`
	ViolationsCount int                 `json:"violations_count"`
	CoveragePercent *float64            `json:"coverage_percent,omitempty"`
	CoverageByFile  []judgeFileCoverage `json:"coverage_by_file,omitempty"`
	CoverageTree    *judgeCoverageNode  `json:"coverage_tree,omitempty"`
	FindingsTree    *judgeFindingsNode  `json:"findings_tree,omitempty"`
	Rulings         []judgeRuling       `json:"rulings"`
	Violations      []judgeViolation    `json:"violations,omitempty"`
	Delta           *judgeDelta         `json:"delta,omitempty"`
}

type judgeCoverageNode struct {
	Path         string              `json:"path"`
	CoveredLines int                 `json:"covered_lines"`
	TotalLines   int                 `json:"total_lines"`
	Percent      float64             `json:"percent"`
	Children     []judgeCoverageNode `json:"children,omitempty"`
	Files        []judgeCoverageFile `json:"files,omitempty"`
}

type judgeCoverageFile struct {
	Name         string  `json:"name"`
	CoveredLines int     `json:"covered_lines"`
	TotalLines   int     `json:"total_lines"`
	Percent      float64 `json:"percent"`
}

type judgeFindingsNode struct {
	Path       string              `json:"path"`
	Count      int                 `json:"count"`
	BySeverity map[string]int      `json:"by_severity,omitempty"`
	Children   []judgeFindingsNode `json:"children,omitempty"`
	Files      []judgeFindingsFile `json:"files,omitempty"`
}

type judgeFindingsFile struct {
	Name       string         `json:"name"`
	Count      int            `json:"count"`
	BySeverity map[string]int `json:"by_severity,omitempty"`
}

type judgeFileCoverage struct {
	FilePath        string   `json:"file_path"`
	CoveredLines    int      `json:"covered_lines"`
	TotalLines      int      `json:"total_lines"`
	Percent         float64  `json:"percent"`
	Covered         []int    `json:"covered"`
	Uncovered       []int    `json:"uncovered"`
	PreviousPercent *float64 `json:"previous_percent,omitempty"`
	CoverageDelta   *float64 `json:"coverage_delta,omitempty"`
	IsNew           bool     `json:"is_new,omitempty"`
}

type judgeViolation struct {
	Rule      string `json:"rule"`
	SourcePkg string `json:"source_pkg"`
	TargetPkg string `json:"target_pkg"`
	Message   string `json:"message"`
	Status    string `json:"status,omitempty"`
}

type judgeRuling struct {
	Subtype string `json:"subtype"`
	Passed  bool   `json:"passed"`
	Detail  string `json:"detail"`
}

type judgeDelta struct {
	HasPrevious bool `json:"has_previous"`
	NewCount    int  `json:"new_count"`
	FixedCount  int  `json:"fixed_count"`
}

func formatJudgeOutput(output []byte, exitCode int) (string, error) {
	var resp judgeResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return string(output), nil
	}

	var builder strings.Builder
	for _, proj := range resp.Projects {
		icon := verdictPass
		if proj.Verdict == "fail" {
			icon = verdictFail
		}
		fmt.Fprintf(&builder, "## %s — %s\n\n", proj.Name, icon)
		fmt.Fprintf(&builder, "Findings: %d\n", proj.FindingsCount)
		if proj.ViolationsCount > 0 {
			fmt.Fprintf(&builder, "Architecture violations: %d\n", proj.ViolationsCount)
			for _, violation := range proj.Violations {
				marker := ""
				if violation.Status == statusNew {
					marker = "  " + statusNewLabel
				}
				fmt.Fprintf(&builder, "  %s: %s → %s%s\n", violation.Rule, violation.SourcePkg, violation.TargetPkg, marker)
				if violation.Message != "" {
					fmt.Fprintf(&builder, "    %s\n", violation.Message)
				}
			}
		}
		if proj.CoveragePercent != nil {
			fmt.Fprintf(&builder, "Coverage: %.1f%%\n", *proj.CoveragePercent)
		}

		if proj.Delta != nil && proj.Delta.HasPrevious {
			fmt.Fprintf(&builder, "Delta: %d new, %d fixed\n", proj.Delta.NewCount, proj.Delta.FixedCount)
		}

		if len(proj.Rulings) > 0 {
			builder.WriteString("\nRulings:\n")
			for _, ruling := range proj.Rulings {
				status := verdictPass
				if !ruling.Passed {
					status = verdictFail
				}
				fmt.Fprintf(&builder, "  %s: %s", ruling.Subtype, status)
				if ruling.Detail != "" {
					fmt.Fprintf(&builder, " — %s", ruling.Detail)
				}
				builder.WriteString("\n")
			}
		}

		if proj.CoverageTree != nil {
			builder.WriteString("\nCoverage by directory:\n")
			writeCoverageTree(&builder, proj.CoverageTree, 1)
		}
		if proj.FindingsTree != nil && proj.FindingsTree.Count > 0 {
			builder.WriteString("\nFindings by directory:\n")
			writeFindingsTree(&builder, proj.FindingsTree, 1)
		}
		builder.WriteString("\n")
	}

	if exitCode == 1 && len(resp.Projects) > 0 {
		builder.WriteString("Quality gate FAILED. Fix the findings above to pass.\n")
	}

	return builder.String(), nil
}
