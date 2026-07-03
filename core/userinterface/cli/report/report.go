package report

import (
	"errors"

	"github.com/spf13/cobra"
)

// NewCommand builds the `gavel report` command. It delivers the verdict cached
// by `gavel judge` to an external sink (GitHub Checks). Flags are bound from
// the clispec-generated Options; the delivery itself is filled in by later
// phases, so the skeleton fails loudly rather than reporting a false success.
func NewCommand() *cobra.Command {
	opts := Options{}
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Deliver a judged verdict to an external sink (GitHub PR checks)",
		Long: `Read the cached verdict written by 'gavel judge' and deliver it to an external
sink. The only sink today is GitHub Checks: a merge-blocking check run with new
findings annotated inline on the pull-request diff.

'gavel report' never re-runs analysis — run 'gavel judge' first. In GitHub
Actions the built-in GITHUB_TOKEN carries the checks:write permission, so
same-repo pull requests need no extra setup.`,
		Example: `  gavel report
  gavel report --project=payments`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(opts)
		},
	}
	RegisterFlags(cmd, &opts)
	return cmd
}

func run(_ Options) error {
	return errors.New("report: delivery not yet implemented")
}
