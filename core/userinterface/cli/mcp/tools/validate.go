package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
)

type ValidateInput struct {
	Gavelspace string `json:"gavelspace,omitempty" jsonschema:"Absolute path to a gavelspace directory (omit to use the directory where Claude Code is running)"`
}

func RegisterValidate(server *mcp.Server, cli *executor.CLI) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gavel_validate",
		Description: "Validate Gavel structural setup — checks that required files exist, .bazelrc includes are in place, and MODULE.bazel has the expected entries.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ValidateInput) (*mcp.CallToolResult, any, error) {
		result, err := runValidate(ctx, cli, input)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}}, nil, nil
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil, nil
	})
}

func runValidate(ctx context.Context, cli *executor.CLI, input ValidateInput) (string, error) {
	output, exitCode, err := cli.RunIn(ctx, input.Gavelspace, "validate")
	if err != nil {
		return "", fmt.Errorf("execute gavel validate: %w", err)
	}

	text := strings.TrimSpace(string(output))
	if exitCode == 0 {
		return "Gavel structure is valid. All required files and includes are in place.", nil
	}
	return fmt.Sprintf("Gavel structure has issues:\n\n%s\n\nRun `gavel init` to fix.", text), nil
}
