package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
)

func RegisterArchitecture(server *mcp.Server, cli *executor.CLI) {
	server.AddResource(&mcp.Resource{
		URI:         "gavel://architecture",
		Name:        "Architecture policy",
		Description: "Layer definitions and deny rules from architecture.yml — what imports are forbidden",
		MIMEType:    "text/plain",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return readArchitecture(ctx, cli)
	})
}

type archOutput struct {
	Architecture *archDetail `json:"architecture,omitempty"`
}

type archDetail struct {
	Layers []archLayer `json:"layers"`
	Rules  []archDeny  `json:"deny_rules"`
}

type archLayer struct {
	Name     string   `json:"name"`
	Patterns []string `json:"patterns"`
}

type archDeny struct {
	Name   string   `json:"name"`
	Source string   `json:"source"`
	Deny   []string `json:"deny"`
}

func readArchitecture(ctx context.Context, cli *executor.CLI) (*mcp.ReadResourceResult, error) {
	output, _, err := cli.Run(ctx, "projects")
	if err != nil {
		return nil, fmt.Errorf("run gavel projects: %w", err)
	}

	var data archOutput
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("parse projects output: %w", err)
	}

	if data.Architecture == nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:  "gavel://architecture",
				Text: "No architecture policy found. Create .gavel/architecture.yml to define layer rules.",
			}},
		}, nil
	}

	var builder strings.Builder
	fmt.Fprintln(&builder, "# Architecture Policy")
	fmt.Fprintln(&builder)

	if len(data.Architecture.Layers) > 0 {
		fmt.Fprintln(&builder, "## Layers")
		for _, l := range data.Architecture.Layers {
			fmt.Fprintf(&builder, "  %s: %s\n", l.Name, strings.Join(l.Patterns, ", "))
		}
		fmt.Fprintln(&builder)
	}

	if len(data.Architecture.Rules) > 0 {
		fmt.Fprintln(&builder, "## Deny Rules")
		for _, r := range data.Architecture.Rules {
			fmt.Fprintf(&builder, "  %s: %s cannot import %s\n", r.Name, r.Source, strings.Join(r.Deny, ", "))
		}
		fmt.Fprintln(&builder)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{URI: "gavel://architecture", Text: builder.String()}},
	}, nil
}
