package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/usegavel/gavel/core/application/gavelspace/loadgavelspace"
)

type WorkspaceResolver func() (string, error)

type outputDTO struct {
	Projects     []projectDTO     `json:"projects"`
	Architecture *architectureDTO `json:"architecture,omitempty"`
}

type projectDTO struct {
	Name      string      `json:"name"`
	Pattern   string      `json:"pattern"`
	Exclude   []string    `json:"exclude,omitempty"`
	Languages []string    `json:"languages"`
	Gate      gateDTO     `json:"quality_gate"`
	Baseline  baselineDTO `json:"baseline"`
}

type gateDTO struct {
	Rules []ruleDTO `json:"rules"`
}

type ruleDTO struct {
	Subtype string `json:"subtype"`
}

type baselineDTO struct {
	FindingsCount   int `json:"findings_count"`
	ViolationsCount int `json:"violations_count"`
}

type architectureDTO struct {
	Layers []layerDTO `json:"layers"`
	Rules  []denyDTO  `json:"deny_rules"`
}

type layerDTO struct {
	Name     string   `json:"name"`
	Patterns []string `json:"patterns"`
}

type denyDTO struct {
	Name   string   `json:"name"`
	Source string   `json:"source"`
	Deny   []string `json:"deny"`
}

func NewCommand(resolveWorkspace WorkspaceResolver, handler *loadgavelspace.Handler) *cobra.Command {
	var configPath string
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "List configured projects with quality gates and baselines",
		Long:  `Read gavel.yaml and baseline data, output project details as JSON.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, configPath, resolveWorkspace, handler)
		},
	}
	cmd.Flags().StringVar(&configPath, "config", "", "Path to gavel.yaml")
	return cmd
}

func run(cmd *cobra.Command, explicitPath string, resolveWorkspace WorkspaceResolver, handler *loadgavelspace.Handler) error {
	workspace, err := resolveWorkspace()
	if err != nil {
		return err
	}

	configPath := resolveConfigPath(explicitPath, workspace)
	q, err := loadgavelspace.NewQuery(configPath, loadgavelspace.WithWorkspace(workspace))
	if err != nil {
		return err
	}
	result, err := handler.Execute(context.Background(), q)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	view := result.View
	var projects []projectDTO
	for _, proj := range view.Projects {
		var rules []ruleDTO
		for _, r := range proj.GateRules {
			rules = append(rules, ruleDTO{Subtype: r.Subtype})
		}
		projects = append(projects, projectDTO{
			Name:      proj.Name,
			Pattern:   proj.TargetPattern,
			Exclude:   proj.ExcludePatterns,
			Languages: proj.Languages,
			Gate:      gateDTO{Rules: rules},
			Baseline: baselineDTO{
				FindingsCount:   proj.Baseline.FingerprintCount,
				ViolationsCount: proj.Baseline.ArchIDCount,
			},
		})
	}

	out := outputDTO{Projects: projects}
	if len(view.Projects) > 0 && view.Projects[0].ArchPolicy != nil {
		activeProj := view.Projects[0].ArchPolicy
		var layers []layerDTO
		for _, l := range activeProj.Layers {
			layers = append(layers, layerDTO{Name: l.Name, Patterns: l.Patterns})
		}
		var denyRules []denyDTO
		for _, r := range activeProj.Rules {
			denyRules = append(denyRules, denyDTO{Name: r.Name, Source: r.Source, Deny: r.Deny})
		}
		out.Architecture = &architectureDTO{Layers: layers, Rules: denyRules}
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func resolveConfigPath(explicit, workspace string) string {
	if explicit != "" {
		return explicit
	}
	candidate := filepath.Join(workspace, ".gavel", "gavel.yaml")
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	candidate = filepath.Join(workspace, "gavel.yaml")
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return filepath.Join(workspace, ".gavel", "gavel.yaml")
}
