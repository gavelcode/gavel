package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
)

type InitInput struct {
	Gavelspace string `json:"gavelspace,omitempty" jsonschema:"Absolute path to a gavelspace directory (omit to use the directory where Claude Code is running)"`
	From       string `json:"from"                 jsonschema:"Path to an existing gavel.yaml to use as input"`
}

func RegisterInit(server *mcp.Server, cli *executor.CLI) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gavel_init",
		Description: "Initialize Gavel in the workspace from an existing gavel.yaml. Installs Bazel aspects, registers tool dependencies in MODULE.bazel, and sets up .bazelrc includes.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input InitInput) (*mcp.CallToolResult, any, error) {
		result, err := runInit(ctx, cli, input)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}}, nil, nil
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil, nil
	})
}

func runInit(ctx context.Context, cli *executor.CLI, input InitInput) (string, error) {
	args := []string{"init", "--from", input.From, "--force"}

	output, exitCode, err := cli.RunIn(ctx, input.Gavelspace, args...)
	if err != nil {
		return "", fmt.Errorf("execute gavel init: %w", err)
	}

	text := strings.TrimSpace(string(output))
	if exitCode != 0 {
		return fmt.Sprintf("gavel init failed:\n\n%s", text), nil
	}
	return text, nil
}
