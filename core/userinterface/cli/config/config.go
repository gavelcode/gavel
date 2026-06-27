package config

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
	ConfigPath string       `json:"config_path"`
	Gavelspace string       `json:"gavelspace,omitempty"`
	Server     string       `json:"server,omitempty"`
	Projects   []projectDTO `json:"projects"`
}

type projectDTO struct {
	Name      string   `json:"name"`
	Pattern   string   `json:"pattern"`
	Exclude   []string `json:"exclude,omitempty"`
	Languages []string `json:"languages"`
	Gate      gateDTO  `json:"quality_gate"`
}

type gateDTO struct {
	Rules []ruleDTO `json:"rules"`
}

type ruleDTO struct {
	Subtype string `json:"subtype"`
}

func NewCommand(resolveWorkspace WorkspaceResolver, handler *loadgavelspace.Handler) *cobra.Command {
	var configPath string
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show workspace configuration",
		Long:  `Read gavel.yaml and output the parsed workspace configuration as JSON.`,
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
	q, err := loadgavelspace.NewQuery(configPath)
	if err != nil {
		return err
	}
	result, err := handler.Execute(context.Background(), q)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	view := result.View
	out := outputDTO{ConfigPath: configPath}
	if view.GavelspaceName != "" {
		out.Gavelspace = view.GavelspaceName
	}
	if view.ServerURL != "" {
		out.Server = view.ServerURL
	}

	for _, proj := range view.Projects {
		var rules []ruleDTO
		for _, r := range proj.GateRules {
			rules = append(rules, ruleDTO{Subtype: r.Subtype})
		}
		out.Projects = append(out.Projects, projectDTO{
			Name:      proj.Name,
			Pattern:   proj.TargetPattern,
			Exclude:   proj.ExcludePatterns,
			Languages: proj.Languages,
			Gate:      gateDTO{Rules: rules},
		})
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
