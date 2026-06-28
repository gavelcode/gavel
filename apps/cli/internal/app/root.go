package app

import (
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	ingestcov "github.com/usegavel/gavel/core/application/casefile/ingestcoverage"
	ingestfind "github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	"github.com/usegavel/gavel/core/application/casefile/submit"
	"github.com/usegavel/gavel/core/application/gavelspace/loadgavelspace"
	"github.com/usegavel/gavel/core/application/project/preparebaseline"
	"github.com/usegavel/gavel/core/application/supporting/analyzetarget"
	"github.com/usegavel/gavel/core/userinterface/cli/config"
	"github.com/usegavel/gavel/core/userinterface/cli/initgavel"
	"github.com/usegavel/gavel/core/userinterface/cli/judge"
	gavelmcp "github.com/usegavel/gavel/core/userinterface/cli/mcp"
	"github.com/usegavel/gavel/core/userinterface/cli/projects"
	"github.com/usegavel/gavel/core/userinterface/cli/trends"
	"github.com/usegavel/gavel/core/userinterface/cli/validate"
	"github.com/usegavel/gavel/core/userinterface/cli/watch"
)

type Deps struct {
	Version string

	WorkspaceResolver func() (string, error)
	Logger            *slog.Logger
	LogLevel          *slog.LevelVar

	FindingsHandler  *ingestfind.Handler
	CoverageHandler  *ingestcov.Handler
	SubmitHandler    *submit.Handler
	CollectEvHandler *collectevidence.Handler
	LoadWsHandler    *loadgavelspace.Handler
	ProjectRepo      preparebaseline.ProjectRepository
	FPSeeder         preparebaseline.FingerprintSeeder
	SourceContext    judge.SourceContext
	TargetQuery      judge.TargetQuery

	Verifier        judge.StructureVerifier
	ConfigInstaller initgavel.ConfigInstaller
	ToolCatalog     initgavel.ToolCatalogProvider

	AnalyzeHandler      *analyzetarget.Handler
	JudgeTargetResolver judge.TargetResolver
	WatchTargetResolver analyzetarget.TargetResolver
}

func NewRootCommand(deps Deps) *cobra.Command {
	var verbose bool
	root := &cobra.Command{
		Use:     "gavel",
		Version: deps.Version,
		Short:   "Gavel CLI — monorepo Bazel quality analyzer",
		Long: `Gavel is a quality gate tool for Bazel monorepos. It runs static analyzers
(PMD, SpotBugs, golangci-lint, ESLint, Clippy, Ruff, Bandit, and more) as Bazel
aspects, collects coverage via bazel coverage, enforces architecture constraints,
and evaluates a configurable quality gate to produce a pass/fail verdict.

Gavel supports Go, Java, Kotlin, Python, TypeScript, and Rust. All analysis runs
through Bazel's action cache for fast incremental builds.

Start with 'gavel init' to set up a workspace, then 'gavel judge' to analyze.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			if verbose {
				deps.LogLevel.Set(slog.LevelDebug)
			}
		},
	}
	root.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable debug logging")

	root.AddCommand(judge.NewCommand(
		deps.FindingsHandler, deps.CoverageHandler,
		deps.SubmitHandler, deps.CollectEvHandler,
		deps.LoadWsHandler, deps.ProjectRepo, deps.FPSeeder,
		deps.WorkspaceResolver, deps.SourceContext, deps.Verifier,
		deps.Logger, deps.TargetQuery, deps.JudgeTargetResolver,
	))
	root.AddCommand(initgavel.NewCommand(deps.WorkspaceResolver, deps.ConfigInstaller, deps.ToolCatalog))
	root.AddCommand(validate.NewCommand(deps.WorkspaceResolver, deps.Verifier))
	root.AddCommand(watch.NewCommand(deps.AnalyzeHandler, deps.WatchTargetResolver))
	root.AddCommand(config.NewCommand(deps.WorkspaceResolver, deps.LoadWsHandler))
	root.AddCommand(projects.NewCommand(deps.WorkspaceResolver, deps.LoadWsHandler))
	root.AddCommand(trends.NewCommand())
	root.AddCommand(gavelmcp.NewCommand(deps.Version))

	return root
}
