package initgavel

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/usegavel/gavel/core/userinterface/cli/ui"
)

type WorkspaceResolver func() (string, error)

const (
	gavelBazelrc = ".gavel/gavel.bazelrc"
	gavelModule  = ".gavel/gavel.MODULE.bazel"
	totalPhases  = 3
)

func NewCommand(resolveWorkspace WorkspaceResolver, inst ConfigInstaller, catalog ToolCatalogProvider) *cobra.Command {
	opts := Options{}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Gavel configuration for this Bazel workspace",
		Long: `Initialize Gavel in the current Bazel workspace. Creates gavel.yaml with
project definitions, registers analysis aspects in .bazelrc, and installs tool
dependencies in MODULE.bazel.

Interactive by default: prompts for the project name and project definitions.
Use --from to re-apply an existing gavel.yaml non-interactively (CI, scripting).`,
		Example: `  gavel init
  gavel init --from=gavel.yaml
  gavel init --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, opts, resolveWorkspace, inst, catalog)
		},
	}
	RegisterFlags(cmd, &opts)
	return cmd
}

func run(cmd *cobra.Command, opts Options, resolveWorkspace WorkspaceResolver, inst ConfigInstaller, catalog ToolCatalogProvider) error {
	writer := cmd.OutOrStdout()
	if _, err := fmt.Fprint(writer, ui.Header("INIT")); err != nil {
		return err
	}

	workspace, err := resolveWorkspace()
	if err != nil {
		return err
	}
	configPath := opts.Config
	if configPath == "" {
		configPath = ".gavel/gavel.yaml"
	}

	var name string
	var projects []Project
	var server Server
	force := opts.Force

	if opts.From != "" {
		name, projects, server, err = readFromConfig(opts.From)
		if err != nil {
			return err
		}
		force = true
	} else {
		defaultName := filepath.Base(workspace)
		if defaultName == "." || defaultName == string(filepath.Separator) {
			defaultName = "gavel-project"
		}
		name, err = promptProjectName(defaultName)
		if err != nil {
			return err
		}
		projects, err = promptProjects()
		if err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(writer, ui.PhaseHeader(1, totalPhases, "CONFIG", "Writing project configuration")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "  %s\n", ui.Dim.Render("Defines project name, Bazel targets, and language settings.")); err != nil {
		return err
	}

	result, err := execute(configPath, workspace, name, force, projects, server, opts.From, inst, catalog)
	if err != nil {
		if _, wErr := fmt.Fprint(writer, ui.PhaseItem(configPath, "FAILED", false)); wErr != nil {
			return wErr
		}
		return err
	}

	if !result.created {
		if _, err := fmt.Fprint(writer, ui.ExistingConfig(configPath)); err != nil {
			return err
		}
		return nil
	}

	if _, err := fmt.Fprint(writer, ui.PhaseItem(result.configPath, "CREATED", true)); err != nil {
		return err
	}

	const phaseBazel = 2
	if _, err := fmt.Fprint(writer, ui.PhaseHeader(phaseBazel, totalPhases, "BAZEL", "Registering analysis aspects")); err != nil {
		return err
	}
	if _, err := fmt.Fprint(writer, ui.PhaseItem(gavelBazelrc, "CREATED", true)); err != nil {
		return err
	}
	if _, err := fmt.Fprint(writer, ui.PhaseItem(".bazelrc", fileStatus(result.modified, ".bazelrc", "UPDATED"), true)); err != nil {
		return err
	}
	for _, aspect := range result.aspects {
		if _, err := fmt.Fprint(writer, ui.PhaseItem(aspect, "ADDED", true)); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(writer, ui.PhaseHeader(totalPhases, totalPhases, "MODULE", "Installing tool dependencies")); err != nil {
		return err
	}
	if _, err := fmt.Fprint(writer, ui.PhaseItem(gavelModule, "CREATED", true)); err != nil {
		return err
	}
	if _, err := fmt.Fprint(writer, ui.PhaseItem("MODULE.bazel", fileStatus(result.modified, "MODULE.bazel", "UPDATED"), true)); err != nil {
		return err
	}
	for _, binary := range result.binaries {
		if _, err := fmt.Fprint(writer, ui.PhaseItem(binary, "ADDED", true)); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(writer, "\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "  %s  %s\n", ui.Dim.Render("Project"), ui.Label.Render(name)); err != nil {
		return err
	}
	if err := printProjectSummary(writer, projects); err != nil {
		return err
	}

	if _, err := fmt.Fprint(writer, ui.Verdict(result.configPath)); err != nil {
		return err
	}
	return nil
}

func fileStatus(modified map[string]bool, name, defaultStatus string) string {
	if modified[name] {
		return defaultStatus
	}
	return "UNCHANGED"
}

func printProjectSummary(writer io.Writer, projects []Project) error {
	for _, p := range projects {
		if _, err := fmt.Fprintf(writer, "  %s  %s  %s  %s\n",
			ui.Dim.Render("Project"),
			ui.Label.Render(p.Name),
			ui.Dim.Render(p.Pattern),
			strings.Join(p.Tooling, ", "),
		); err != nil {
			return err
		}
	}
	return nil
}
