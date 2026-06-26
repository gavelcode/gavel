package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
)

const defaultFindingsLimit = 100

type FindingsInput struct {
	Gavelspace string `json:"gavelspace,omitempty" jsonschema:"Absolute path to a gavelspace directory (omit to use the directory where Claude Code is running)"`
	Project    string `json:"project,omitempty"    jsonschema:"Analyze only this project (omit for all projects)"`
	Rule       string `json:"rule,omitempty"       jsonschema:"Keep only findings whose rule id matches (omit for all rules)"`
	Severity   string `json:"severity,omitempty"   jsonschema:"Keep only findings of this severity, e.g. error or warning (omit for all)"`
	Limit      int    `json:"limit,omitempty"      jsonschema:"Maximum findings to list in detail; the by-rule summary always counts all matches (default 100)"`
}

func RegisterFindings(server *mcp.Server, cli *executor.CLI) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gavel_findings",
		Description: "Discover every lint finding across a project in one call. Runs a quick analysis (findings only, no coverage), returns a by-rule summary plus a flat file:line list with gate-blocking NEW findings sorted first. Filter by project, rule, or severity.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input FindingsInput) (*mcp.CallToolResult, any, error) {
		result, err := runFindings(ctx, cli, input)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}}, nil, nil
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil, nil
	})
}

func runFindings(ctx context.Context, cli *executor.CLI, input FindingsInput) (string, error) {
	args := []string{"judge", "--quick"}
	if input.Project != "" {
		args = append(args, "--project", input.Project)
	}

	output, _, err := cli.RunInJSON(ctx, input.Gavelspace, args...)
	if err != nil {
		return "", fmt.Errorf("execute gavel judge: %w", err)
	}

	var resp lintResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return "", fmt.Errorf("parse judge output: %w", err)
	}

	filtered := filterFindings(resp, input.Rule, input.Severity)
	if len(filtered) == 0 {
		return "No findings.", nil
	}

	sortFindings(filtered)

	limit := input.Limit
	if limit <= 0 {
		limit = defaultFindingsLimit
	}
	shown := filtered
	if len(shown) > limit {
		shown = shown[:limit]
	}

	return renderFindings(filtered, shown, limit), nil
}

func filterFindings(resp lintResponse, rule, severity string) []lintFinding {
	var filtered []lintFinding
	for _, project := range resp.Projects {
		for _, finding := range project.Findings {
			if rule != "" && finding.RuleID != rule {
				continue
			}
			if severity != "" && finding.Severity != severity {
				continue
			}
			filtered = append(filtered, finding)
		}
	}
	return filtered
}

func sortFindings(findings []lintFinding) {
	sort.SliceStable(findings, func(i, j int) bool {
		left, right := findings[i], findings[j]
		if isNew(left) != isNew(right) {
			return isNew(left)
		}
		if left.FilePath != right.FilePath {
			return left.FilePath < right.FilePath
		}
		return left.Line < right.Line
	})
}

func isNew(f lintFinding) bool {
	return f.Status == statusNew
}

func renderFindings(filtered, shown []lintFinding, limit int) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# Findings (%d)\n\n", len(filtered))
	fmt.Fprintf(&builder, "By rule: %s\n\n", summarizeByRule(filtered))

	for _, f := range shown {
		fmt.Fprintf(&builder, "%s:%d [%s] %s (%s:%s)", f.FilePath, f.Line, f.Severity, f.Message, f.Tool, f.RuleID)
		if isNew(f) {
			builder.WriteString(" " + statusNewLabel)
		}
		builder.WriteString("\n")
	}

	if len(filtered) > limit {
		fmt.Fprintf(&builder, "\nShowing %d of %d. Narrow with rule= or severity=, or raise limit=.\n", limit, len(filtered))
	}
	return builder.String()
}

func summarizeByRule(findings []lintFinding) string {
	counts := map[string]int{}
	var order []string
	for _, f := range findings {
		if counts[f.RuleID] == 0 {
			order = append(order, f.RuleID)
		}
		counts[f.RuleID]++
	}
	sort.SliceStable(order, func(i, j int) bool {
		if counts[order[i]] != counts[order[j]] {
			return counts[order[i]] > counts[order[j]]
		}
		return order[i] < order[j]
	})

	parts := make([]string, len(order))
	for i, rule := range order {
		parts[i] = fmt.Sprintf("%d %s", counts[rule], rule)
	}
	return strings.Join(parts, ", ")
}
