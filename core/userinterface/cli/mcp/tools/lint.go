package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
)

type LintFileInput struct {
	Gavelspace string `json:"gavelspace,omitempty" jsonschema:"Absolute path to a gavelspace directory (omit to use the directory where Claude Code is running)"`
	File       string `json:"file"                 jsonschema:"File path relative to workspace root"`
	Project    string `json:"project,omitempty"     jsonschema:"Project name (analyzes all projects if omitted)"`
}

func RegisterLintFile(server *mcp.Server, cli *executor.CLI) {
	mcp.AddTool(server, &mcp.Tool{
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
		Name:        "gavel_lint_file",
		Description: "Get lint findings for a specific file. Runs a quick analysis (findings only, no coverage) and returns findings filtered to the requested file path.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input LintFileInput) (*mcp.CallToolResult, any, error) {
		result, err := runLintFile(ctx, cli, input)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}}, nil, nil
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil, nil
	})
}

type lintFinding struct {
	Tool        string `json:"tool"`
	RuleID      string `json:"rule_id"`
	Severity    string `json:"severity"`
	FilePath    string `json:"file_path"`
	Line        int    `json:"line"`
	Message     string `json:"message"`
	Fingerprint string `json:"fingerprint"`
	Status      string `json:"status,omitempty"`
}

type lintResponse struct {
	Projects []lintProject `json:"projects"`
}

type lintProject struct {
	Name     string        `json:"name"`
	Findings []lintFinding `json:"findings"`
}

func runLintFile(ctx context.Context, cli *executor.CLI, input LintFileInput) (string, error) {
	args := []string{"judge", "--quick", "--no-baseline-update", "--target-file", input.File}
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

	var matched []lintFinding
	for _, p := range resp.Projects {
		for _, f := range p.Findings {
			if f.FilePath == input.File {
				matched = append(matched, f)
			}
		}
	}

	if len(matched) == 0 {
		return fmt.Sprintf("No findings in %s", input.File), nil
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "# Findings in %s (%d)\n\n", input.File, len(matched))
	for _, f := range matched {
		fmt.Fprintf(&builder, "Line %d: [%s] %s (%s:%s)", f.Line, f.Severity, f.Message, f.Tool, f.RuleID)
		if f.Status == statusNew {
			builder.WriteString(" " + statusNewLabel)
		}
		builder.WriteString("\n")
	}
	return builder.String(), nil
}
