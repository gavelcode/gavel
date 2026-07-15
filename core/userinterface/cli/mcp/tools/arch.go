package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
)

type ArchInput struct {
	Gavelspace string `json:"gavelspace,omitempty" jsonschema:"Absolute path to a gavelspace directory (omit to use the directory where Claude Code is running)"`
	Project    string `json:"project,omitempty"    jsonschema:"Analyze only this project (omit for all projects)"`
}

func RegisterArch(server *mcp.Server, cli *executor.CLI) {
	mcp.AddTool(server, &mcp.Tool{
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
		Name:        "gavel_arch",
		Description: "Check architecture layer violations (DDD layer rules). Returns per-project violations with rule, source and target packages. Runs a full analysis (not --quick) because architecture checks are skipped in quick mode.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ArchInput) (*mcp.CallToolResult, any, error) {
		result, err := runArch(ctx, cli, input)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}}, nil, nil
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil, nil
	})
}

func runArch(ctx context.Context, cli *executor.CLI, input ArchInput) (string, error) {
	args := []string{"judge", "--no-baseline-update"}
	if input.Project != "" {
		args = append(args, "--project", input.Project)
	}

	output, _, err := cli.RunInJSON(ctx, input.Gavelspace, args...)
	if err != nil {
		return "", fmt.Errorf("execute gavel judge: %w", err)
	}

	return formatArchOutput(output)
}

func formatArchOutput(output []byte) (string, error) {
	var resp judgeResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return "", fmt.Errorf("parse judge output: %w", err)
	}

	totalViolations := 0
	for _, proj := range resp.Projects {
		totalViolations += proj.ViolationsCount
	}

	if totalViolations == 0 {
		return "No architecture violations found.", nil
	}

	var builder strings.Builder
	for _, proj := range resp.Projects {
		if proj.ViolationsCount == 0 {
			continue
		}
		fmt.Fprintf(&builder, "## %s — %d violations\n\n", proj.Name, proj.ViolationsCount)
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
		builder.WriteString("\n")
	}

	return builder.String(), nil
}
