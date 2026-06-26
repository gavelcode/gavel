package mcp

import (
	"os"

	"github.com/spf13/cobra"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
	"github.com/usegavel/gavel/core/userinterface/cli/mcp/resources"
	"github.com/usegavel/gavel/core/userinterface/cli/mcp/tools"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server for LLM agent integration",
		Long: `Start a Model Context Protocol (MCP) server on stdio. Exposes Gavel
resources (configuration, projects, quality gates, baselines, architecture
policy) and tools (judge, findings, lint-file, coverage, validate) to MCP-compatible
LLM agents such as Claude Code, Cursor, and VS Code Copilot.

Configure in Claude Code settings.json:
  { "mcpServers": { "gavel": { "command": "gavel", "args": ["mcp"] } } }`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd)
		},
	}
	return cmd
}

func run(cmd *cobra.Command) error {
	workspace := os.Getenv("GAVEL_WORKSPACE")
	cli := executor.New(workspace)
	server := NewServer(cli)
	return server.Run(cmd.Context(), &mcpsdk.StdioTransport{})
}

func NewServer(cli *executor.CLI) *mcpsdk.Server {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "gavel",
		Version: "0.1.0",
	}, nil)

	resources.RegisterConfig(server, cli)
	resources.RegisterProjects(server, cli)
	resources.RegisterArchitecture(server, cli)

	tools.RegisterInit(server, cli)
	tools.RegisterJudge(server, cli)
	tools.RegisterFindings(server, cli)
	tools.RegisterLintFile(server, cli)
	tools.RegisterCoverage(server, cli)
	tools.RegisterValidate(server, cli)
	tools.RegisterTrends(server, cli)
	tools.RegisterArch(server, cli)

	return server
}
