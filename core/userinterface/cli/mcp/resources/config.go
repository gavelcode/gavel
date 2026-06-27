package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
)

func RegisterConfig(server *mcp.Server, cli *executor.CLI) {
	server.AddResource(&mcp.Resource{
		URI:         "gavel://config",
		Name:        "Gavel configuration",
		Description: "Workspace configuration from gavel.yaml — projects, quality gates, languages, server settings",
		MIMEType:    "text/plain",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return readConfig(ctx, cli)
	})
}

type configOutput struct {
	ConfigPath string          `json:"config_path"`
	Gavelspace string          `json:"gavelspace,omitempty"`
	Server     string          `json:"server,omitempty"`
	Projects   []configProject `json:"projects"`
}

type configProject struct {
	Name      string     `json:"name"`
	Pattern   string     `json:"pattern"`
	Exclude   []string   `json:"exclude,omitempty"`
	Languages []string   `json:"languages"`
	Gate      configGate `json:"quality_gate"`
}

type configGate struct {
	Rules []configRule `json:"rules"`
}

type configRule struct {
	Subtype string `json:"subtype"`
}

func readConfig(ctx context.Context, cli *executor.CLI) (*mcp.ReadResourceResult, error) {
	output, _, err := cli.Run(ctx, "config")
	if err != nil {
		return nil, fmt.Errorf("run gavel config: %w", err)
	}

	var cfg configOutput
	if err := json.Unmarshal(output, &cfg); err != nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{URI: "gavel://config", Text: string(output)}},
		}, nil
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "# Gavel Configuration (%s)\n\n", cfg.ConfigPath)
	if cfg.Gavelspace != "" {
		fmt.Fprintf(&builder, "Gavelspace: %s\n", cfg.Gavelspace)
	}
	fmt.Fprintf(&builder, "Projects: %d\n\n", len(cfg.Projects))
	for _, proj := range cfg.Projects {
		fmt.Fprintf(&builder, "## %s\n", proj.Name)
		fmt.Fprintf(&builder, "  Pattern: %s\n", proj.Pattern)
		if len(proj.Exclude) > 0 {
			fmt.Fprintf(&builder, "  Exclude: %s\n", strings.Join(proj.Exclude, ", "))
		}
		if len(proj.Languages) > 0 {
			fmt.Fprintf(&builder, "  Languages: %s\n", strings.Join(proj.Languages, ", "))
		}
		if len(proj.Gate.Rules) > 0 {
			fmt.Fprintf(&builder, "  Quality Gate: %d rules\n", len(proj.Gate.Rules))
			for _, r := range proj.Gate.Rules {
				fmt.Fprintf(&builder, "    - %s\n", r.Subtype)
			}
		}
		fmt.Fprintln(&builder)
	}
	if cfg.Server != "" {
		fmt.Fprintf(&builder, "Server: %s\n", cfg.Server)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{URI: "gavel://config", Text: builder.String()}},
	}, nil
}
