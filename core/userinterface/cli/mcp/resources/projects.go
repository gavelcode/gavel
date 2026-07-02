package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
)

func RegisterProjects(server *mcp.Server, cli *executor.CLI) {
	server.AddResource(&mcp.Resource{
		URI:         "gavel://projects",
		Name:        "Gavel projects",
		Description: "All projects configured in gavel.yaml with their patterns, languages, and quality gate summaries",
		MIMEType:    "text/plain",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return readProjects(ctx, cli)
	})

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "gavel://projects/{name}/quality-gate",
		Name:        "Project quality gate",
		Description: "Quality gate rules for a specific project — subtypes and thresholds",
		MIMEType:    "text/plain",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		name := extractProjectName(req.Params.URI, "/quality-gate")
		return readQualityGate(ctx, cli, name, req.Params.URI)
	})

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "gavel://projects/{name}/baseline",
		Name:        "Project baseline",
		Description: "Current baseline state — fingerprint and architecture violation counts",
		MIMEType:    "text/plain",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		name := extractProjectName(req.Params.URI, "/baseline")
		return readBaseline(ctx, cli, name, req.Params.URI)
	})
}

type projectsOutput struct {
	Projects []projectEntry `json:"projects"`
}

type projectEntry struct {
	Name      string      `json:"name"`
	Pattern   string      `json:"pattern"`
	Languages []string    `json:"languages"`
	Gate      projectGate `json:"quality_gate"`
	Baseline  projectBase `json:"baseline"`
}

type projectGate struct {
	Rules []projectRule `json:"rules"`
}

type projectRule struct {
	Subtype string `json:"subtype"`
}

type projectBase struct {
	FindingsCount   int `json:"findings_count"`
	ViolationsCount int `json:"violations_count"`
}

func fetchProjects(ctx context.Context, cli *executor.CLI) (projectsOutput, error) {
	output, _, err := cli.Run(ctx, "projects")
	if err != nil {
		return projectsOutput{}, fmt.Errorf("run gavel projects: %w", err)
	}
	var out projectsOutput
	if err := json.Unmarshal(output, &out); err != nil {
		return projectsOutput{}, fmt.Errorf("parse projects output: %w", err)
	}
	return out, nil
}

func readProjects(ctx context.Context, cli *executor.CLI) (*mcp.ReadResourceResult, error) {
	data, err := fetchProjects(ctx, cli)
	if err != nil {
		return nil, err
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "# Projects (%d)\n\n", len(data.Projects))
	for _, proj := range data.Projects {
		fmt.Fprintf(&builder, "## %s\n", proj.Name)
		fmt.Fprintf(&builder, "  Pattern: %s\n", proj.Pattern)
		if len(proj.Languages) > 0 {
			fmt.Fprintf(&builder, "  Languages: %s\n", strings.Join(proj.Languages, ", "))
		}
		fmt.Fprintf(&builder, "  Quality gate: %d rules\n", len(proj.Gate.Rules))
		fmt.Fprintln(&builder)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{URI: "gavel://projects", Text: builder.String()}},
	}, nil
}

func readQualityGate(ctx context.Context, cli *executor.CLI, projectName, uri string) (*mcp.ReadResourceResult, error) {
	data, err := fetchProjects(ctx, cli)
	if err != nil {
		return nil, err
	}

	for _, proj := range data.Projects {
		if proj.Name != projectName {
			continue
		}
		var builder strings.Builder
		fmt.Fprintf(&builder, "# Quality Gate — %s\n\n", proj.Name)
		if len(proj.Gate.Rules) == 0 {
			fmt.Fprintln(&builder, "No quality gate rules configured.")
		}
		for _, r := range proj.Gate.Rules {
			fmt.Fprintf(&builder, "- %s\n", r.Subtype)
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{URI: uri, Text: builder.String()}},
		}, nil
	}

	return nil, fmt.Errorf("project %q not found", projectName)
}

func readBaseline(ctx context.Context, cli *executor.CLI, projectName, uri string) (*mcp.ReadResourceResult, error) {
	data, err := fetchProjects(ctx, cli)
	if err != nil {
		return nil, err
	}

	for _, proj := range data.Projects {
		if proj.Name != projectName {
			continue
		}
		var builder strings.Builder
		fmt.Fprintf(&builder, "# Baseline — %s\n\n", proj.Name)
		if proj.Baseline.FindingsCount == 0 && proj.Baseline.ViolationsCount == 0 {
			fmt.Fprintln(&builder, "No baseline data found. Run `gavel judge` to establish a baseline.")
		} else {
			fmt.Fprintf(&builder, "Findings fingerprints: %d\n", proj.Baseline.FindingsCount)
			fmt.Fprintf(&builder, "Architecture violations: %d\n", proj.Baseline.ViolationsCount)
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{URI: uri, Text: builder.String()}},
		}, nil
	}

	return nil, fmt.Errorf("project %q not found", projectName)
}

func extractProjectName(uri, suffix string) string {
	uri = strings.TrimPrefix(uri, "gavel://projects/")
	uri = strings.TrimSuffix(uri, suffix)
	return uri
}
