package validate

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/usegavel/gavel/core/userinterface/cli/ui"
)

type WorkspaceResolver func() (string, error)

type StructureVerifier interface {
	VerifyStructure(workspace string) ([]string, error)
}

func NewCommand(resolveWorkspace WorkspaceResolver, verifier StructureVerifier) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate Gavel structural setup (files, includes)",
		Long: `Check that the Gavel structural setup is correct: required files exist,
.bazelrc includes are in place, and MODULE.bazel has the expected entries.

Run this after 'gavel init' or when troubleshooting a broken setup.`,
		Example: `  gavel validate`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, resolveWorkspace, verifier)
		},
	}
	return cmd
}

func run(cmd *cobra.Command, resolveWorkspace WorkspaceResolver, verifier StructureVerifier) error {
	writer := cmd.OutOrStdout()

	workspace, err := resolveWorkspace()
	if err != nil {
		return err
	}

	issues, err := verifier.VerifyStructure(workspace)
	if err != nil {
		return fmt.Errorf("verify structure: %w", err)
	}

	if len(issues) == 0 {
		if _, err := fmt.Fprintf(writer, "%s  Gavel structure is valid\n", ui.Success.Render("VALID")); err != nil {
			return err
		}
		return nil
	}

	if _, err := fmt.Fprintf(writer, "%s  Gavel structure has issues\n", ui.Error.Render("INVALID")); err != nil {
		return err
	}
	for _, issue := range issues {
		if _, err := fmt.Fprintf(writer, "  %s %s\n", ui.Dim.Render("·"), issue); err != nil {
			return err
		}
	}
	return fmt.Errorf("validation failed")
}
