package report

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	outputjson "github.com/usegavel/gavel/core/userinterface/cli/judge/output/json"
	"github.com/usegavel/gavel/core/userinterface/cli/report/checks"
	"github.com/usegavel/gavel/core/userinterface/cli/report/github"
)

const sinkGitHubChecks = "github-checks"

// WorkspaceResolver locates the workspace root that holds .gavel/results.
type WorkspaceResolver func() (string, error)

// ChecksPublisher delivers a built check run to an external sink.
type ChecksPublisher interface {
	Publish(ctx context.Context, checkRun checks.CheckRun) (github.Result, error)
}

// PublisherFactory builds a ChecksPublisher from resolved credentials and target.
type PublisherFactory func(config github.Config) (ChecksPublisher, error)

// NewCommand builds the `gavel report` command. It reads the verdict cached by
// `gavel judge` and delivers it to an external sink (GitHub Checks).
func NewCommand(resolveWorkspace WorkspaceResolver, newPublisher PublisherFactory) *cobra.Command {
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), cmd.OutOrStdout(), opts, resolveWorkspace, newPublisher)
		},
	}
	RegisterFlags(cmd, &opts)
	return cmd
}

func run(ctx context.Context, writer io.Writer, opts Options, resolveWorkspace WorkspaceResolver, newPublisher PublisherFactory) error {
	if opts.To != sinkGitHubChecks {
		return fmt.Errorf("unsupported sink %q (only %q is supported)", opts.To, sinkGitHubChecks)
	}

	workspace, err := resolveWorkspace()
	if err != nil {
		return err
	}

	verdicts, err := outputjson.Load(workspace)
	if errors.Is(err, outputjson.ErrNoResults) {
		return errors.New("nothing to report: run `gavel judge` first")
	}
	if err != nil {
		return err
	}

	if opts.Project != "" {
		verdicts = filterByProject(verdicts, opts.Project)
		if len(verdicts) == 0 {
			return fmt.Errorf("no cached verdict for project %q", opts.Project)
		}
	}

	checkRun := checks.Build(verdicts, checks.Options{
		CheckName: opts.CheckName,
		HeadSHA:   opts.Commit,
		NewOnly:   opts.NewOnly,
	})
	if checkRun.HeadSHA == "" {
		return errors.New("no commit SHA to attach the check run to: pass --commit, " +
			"or run `gavel judge` where it can detect the commit")
	}

	publisher, err := newPublisher(github.Config{Token: opts.GithubToken, Repo: opts.Repo})
	if err != nil {
		return err
	}

	result, err := publisher.Publish(ctx, checkRun)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(writer, "Reported %d project(s) to %s: %s\n",
		len(verdicts), opts.To, result.URL); err != nil {
		return err
	}
	return nil
}

// NewGitHubPublisher is the default PublisherFactory: it builds a GitHub Checks
// publisher from the resolved config. The explicit nil return on error avoids
// handing back a non-nil interface wrapping a nil *github.Publisher.
func NewGitHubPublisher(config github.Config) (ChecksPublisher, error) {
	publisher, err := github.NewPublisher(config)
	if err != nil {
		return nil, err
	}
	return publisher, nil
}

func filterByProject(verdicts []outputjson.Verdict, project string) []outputjson.Verdict {
	filtered := make([]outputjson.Verdict, 0, len(verdicts))
	for _, verdict := range verdicts {
		if verdict.Name == project {
			filtered = append(filtered, verdict)
		}
	}
	return filtered
}
