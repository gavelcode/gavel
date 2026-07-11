package judge

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	apiclient "github.com/usegavel/gavel/core/userinterface/api/v1/client"
	outputjson "github.com/usegavel/gavel/core/userinterface/cli/judge/output/json"
	"github.com/usegavel/gavel/core/userinterface/cli/judge/output/render"
	"github.com/usegavel/gavel/core/userinterface/cli/judge/output/sarif"
	"github.com/usegavel/gavel/core/userinterface/cli/judge/pipeline"
	"github.com/usegavel/gavel/core/userinterface/cli/ui"

	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	ingestcov "github.com/usegavel/gavel/core/application/casefile/ingestcoverage"
	ingestfind "github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	"github.com/usegavel/gavel/core/application/casefile/submit"
	"github.com/usegavel/gavel/core/application/gavelspace/loadgavelspace"
	"github.com/usegavel/gavel/core/application/project/preparebaseline"
)

var ErrVerdictFail = errors.New("one or more projects failed the quality gate")

func NewCommand(
	findings *ingestfind.Handler,
	coverage *ingestcov.Handler,
	submitHandler *submit.Handler,
	collectEvHandler *collectevidence.Handler,
	loadWsHandler *loadgavelspace.Handler,
	projectRepo preparebaseline.ProjectRepository,
	fpSeeder preparebaseline.FingerprintSeeder,
	resolveWorkspace WorkspaceResolver,
	source SourceContext,
	verifier StructureVerifier,
	logger *slog.Logger,
	targetQuery TargetQuery,
	targetResolver TargetResolver,
) *cobra.Command {
	if resolveWorkspace == nil {
		panic("judge: WorkspaceResolver must not be nil")
	}
	if verifier == nil {
		panic("judge: StructureVerifier must not be nil")
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	opts := Options{}
	deps := deps{
		findings:         findings,
		coverage:         coverage,
		submitH:          submitHandler,
		collectEvH:       collectEvHandler,
		loadWorkspace:    loadWsHandler,
		projectRepo:      projectRepo,
		fpSeeder:         fpSeeder,
		resolveWorkspace: resolveWorkspace,
		source:           source,
		validate:         verifier,
		log:              logger,
		targetQuery:      targetQuery,
		targetResolver:   targetResolver,
	}

	cmd := &cobra.Command{
		Use:   "judge",
		Short: "Run analyzers and evaluate quality gate",
		Long: `Run static analyzers as Bazel aspects, collect coverage, check architecture
constraints, and evaluate the quality gate for each project defined in gavel.yaml.

The pipeline for each project:
  1. Run lint aspects (per-language: golangci-lint, PMD, SpotBugs, etc.)
  2. Collect coverage via bazel coverage
  3. Check architecture constraints via archtest aspects
  4. Submit evidence, evaluate quality gate, render verdict

By default, the quality gate evaluates only NEW findings (compared against the
baseline in .gavel/baseline/). Use --absolute to evaluate all findings regardless.
Steps 2 and 3 are skipped with --quick. Use --project to analyze a single project.

Findings can be collected from Gavel's own aspects (default) or from pre-existing
rules_lint reports in bazel-bin/. Use --findings-source to control the mode:
  auto       - detect rules_lint reports, fall back to Gavel aspects (default)
  gavel      - always run Gavel's own Bazel aspects
  rules_lint - read rules_lint reports, supplement with Gavel-exclusive tools

Exit code 0 means all projects passed. Exit code 1 means one or more failed.`,
		Example: `  gavel judge
  gavel judge --project=payments
  gavel judge --quick
  gavel judge --absolute
  gavel judge --json
  gavel judge --timeout=10m
  gavel judge --findings-source=rules_lint
  gavel judge --output-sarif report.sarif`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd, opts, deps)
		},
	}
	RegisterFlags(cmd, &opts)
	return cmd
}

func run(cmd *cobra.Command, opts Options, deps deps) error {
	if opts.Affected || opts.TargetFile != "" {
		opts.Quick = true
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
	defer cancel()
	writer := cmd.OutOrStdout()

	workspace, view, err := setupWorkspace(ctx, writer, &opts, deps)
	if err != nil {
		return err
	}

	if opts.ServerURL == "" && view.ServerURL != "" {
		opts.ServerURL = view.ServerURL
		opts.ServerToken = view.ServerToken
	}
	if opts.ServerURL != "" {
		sc, scErr := apiclient.New(opts.ServerURL, opts.ServerToken)
		if scErr != nil {
			return scErr
		}
		deps.serverClient = sc
		deps.log.Debug("server mode enabled", "url", opts.ServerURL)
	}

	commitSHA, branch, err := resolveGitInfo(ctx, deps.source, opts.CommitSHA, opts.Branch)
	if err != nil {
		return err
	}
	deps.log.Debug("resolved git info", "commit", commitSHA, "branch", branch)

	if !opts.Absolute {
		blResult := prepareBaselines(ctx, deps, view.TenantID, view.Projects)
		if !opts.JSONOutput {
			if err := printBaselineStatus(writer, blResult); err != nil {
				return err
			}
		}
	}
	if !opts.JSONOutput {
		if _, err := fmt.Fprint(writer, ui.Header("JUDGE")); err != nil {
			return err
		}
	}
	if opts.FindingsSource == "" {
		opts.FindingsSource = view.FindingsSource
	}

	startedAt := time.Now().UTC()
	results, err := executeProjects(ctx, writer, deps, workspace, view.TenantID, view.Projects, commitSHA, branch, startedAt, opts)
	if err != nil {
		return err
	}

	if err := emitResults(writer, results, opts, startedAt); err != nil {
		return err
	}
	if err := outputjson.WriteCache(workspace, results); err != nil {
		return fmt.Errorf("write results cache: %w", err)
	}
	if err := emitSARIF(writer, results, opts); err != nil {
		return err
	}

	if !opts.JSONOutput && len(results) > 1 {
		summaries := make([]ui.ProjectSummary, 0, len(results))
		for _, projResult := range results {
			summaries = append(summaries, ui.ProjectSummary{
				Name:        projResult.Name,
				Verdict:     projResult.Verdict,
				Findings:    projResult.FindingsCount,
				NewFindings: projResult.Delta.NewCount,
				Coverage:    projResult.CoveragePercent,
				Violations:  projResult.ViolationsCount,
			})
		}
		if _, err := fmt.Fprint(writer, ui.SummaryTable(summaries, time.Since(startedAt))); err != nil {
			return err
		}
	}

	if hasFailedVerdict(results) {
		return ErrVerdictFail
	}
	return nil
}

func setupWorkspace(ctx context.Context, writer io.Writer, opts *Options, deps deps) (string, loadgavelspace.WorkspaceView, error) {
	workspace, err := deps.resolveWorkspace()
	if err != nil {
		return "", loadgavelspace.WorkspaceView{}, err
	}
	deps.log.Debug("resolved workspace", "path", workspace)

	if err := validateStructure(writer, deps.validate, workspace); err != nil {
		return "", loadgavelspace.WorkspaceView{}, err
	}

	configPath := resolveConfigPath(opts.Config, workspace)
	deps.log.Debug("loading config", "path", configPath)

	var qOpts []loadgavelspace.QueryOption
	qOpts = append(qOpts, loadgavelspace.WithWorkspace(workspace))
	if opts.Project != "" {
		qOpts = append(qOpts, loadgavelspace.WithProjectFilter(opts.Project))
	}
	q, err := loadgavelspace.NewQuery(configPath, qOpts...)
	if err != nil {
		return "", loadgavelspace.WorkspaceView{}, err
	}
	wsResult, err := deps.loadWorkspace.Execute(ctx, q)
	if err != nil {
		return "", loadgavelspace.WorkspaceView{}, fmt.Errorf("load config: %writer", err)
	}

	view := wsResult.View
	if opts.Gavelspace == "" && view.GavelspaceName != "" {
		opts.Gavelspace = view.GavelspaceName
	}

	deps.log.Debug("loaded projects", "count", len(view.Projects))
	return workspace, view, nil
}

func executeProjects(
	ctx context.Context,
	writer io.Writer,
	deps deps,
	workspace string,
	tenantID string,
	projects []loadgavelspace.ProjectView,
	commitSHA, branch string,
	startedAt time.Time,
	opts Options,
) ([]pipeline.Result, error) {
	interactive := !opts.JSONOutput && ui.IsTerminal(writer)
	results := make([]pipeline.Result, 0, len(projects))
	for i, project := range projects {
		if !opts.JSONOutput {
			if _, err := fmt.Fprintf(writer, "\n  %s %s (%d/%d)\n", ui.GoldBar.Render("▸"), project.Name, i+1, len(projects)); err != nil {
				return nil, err
			}
		}
		projResult, err := runProject(ctx, writer, deps, workspace, tenantID, project, commitSHA, branch, startedAt, opts, interactive)
		if err != nil {
			return nil, fmt.Errorf("project %s: %writer", project.Name, err)
		}
		results = append(results, projResult)
	}
	return results, nil
}

func emitResults(writer io.Writer, results []pipeline.Result, opts Options, startedAt time.Time) error {
	if opts.JSONOutput {
		return outputjson.Write(writer, results)
	}
	elapsed := time.Since(startedAt)
	for _, projResult := range results {
		if _, err := fmt.Fprint(writer, render.Findings(projResult)); err != nil {
			return err
		}
		if _, err := fmt.Fprint(writer, render.CoverageSummary(projResult)); err != nil {
			return err
		}
		if _, err := fmt.Fprint(writer, ui.JudgeVerdict(projResult.Verdict, projResult.Name, projResult.FindingsCount, projResult.ViolationsCount, projResult.CoveragePercent, projResult.CoverageSkipped, elapsed)); err != nil {
			return err
		}
		if _, err := fmt.Fprint(writer, ui.DeltaSummary(projResult.FindingsCount, projResult.Delta.FindingsDelta, projResult.CoveragePercent, projResult.Delta.CoverageDelta, projResult.Delta.NewCount, projResult.Delta.FixedCount, projResult.Delta.ExistingCount, projResult.Delta.HasPrevious, projResult.Delta.NewViolationsCount, projResult.Delta.FixedViolationsCount, projResult.Delta.ExistingViolationsCount, projResult.Delta.HasArchPrevious)); err != nil {
			return err
		}
		for i, rl := range projResult.Rulings {
			last := i == len(projResult.Rulings)-1
			if _, err := fmt.Fprint(writer, ui.RulingLine(rl.Subtype, rl.Passed, rl.Detail, last)); err != nil {
				return err
			}
		}
		if projResult.FirstRun {
			if _, err := fmt.Fprint(writer, ui.FirstRunHint()); err != nil {
				return err
			}
		}
		if projResult.ServerFailed {
			if _, err := fmt.Fprint(writer, ui.ServerFallbackWarning()); err != nil {
				return err
			}
		}
		if projResult.BuildWarning != "" {
			if _, err := fmt.Fprint(writer, ui.BuildWarning()); err != nil {
				return err
			}
		}
		if len(projResult.UnanalyzedTools) > 0 {
			if _, err := fmt.Fprint(writer, ui.MissingTargetsWarning(projResult.Name, projResult.UnanalyzedTools)); err != nil {
				return err
			}
		}
	}
	return nil
}

func emitSARIF(writer io.Writer, results []pipeline.Result, opts Options) error {
	if opts.OutputSARIF == "" {
		return nil
	}
	docs := make([][]byte, 0)
	for _, r := range results {
		docs = append(docs, r.RawSARIFDocs...)
	}
	if err := sarif.Write(opts.OutputSARIF, docs); err != nil {
		return fmt.Errorf("write SARIF output: %writer", err)
	}
	if !opts.JSONOutput {
		if _, err := fmt.Fprintf(writer, "\n  SARIF report written to %s\n", opts.OutputSARIF); err != nil {
			return err
		}
	}
	return nil
}

func runProject(
	ctx context.Context,
	writer io.Writer,
	deps deps,
	workspace string,
	tenantID string,
	project loadgavelspace.ProjectView,
	commitSHA, branch string,
	startedAt time.Time,
	opts Options,
	interactive bool,
) (pipeline.Result, error) {
	var cmdOpts []collectevidence.CommandOption
	if opts.TargetFile != "" && deps.targetResolver != nil {
		ownerTarget, ownerErr := deps.targetResolver.FindOwnerTarget(ctx, workspace, opts.TargetFile)
		if ownerErr != nil {
			return pipeline.Result{}, fmt.Errorf("find owner target: %writer", ownerErr)
		}
		cmdOpts = append(cmdOpts, collectevidence.WithScopedTargets([]string{ownerTarget}))
		deps.log.Debug("target-file analysis", "file", opts.TargetFile, "target", ownerTarget)
	} else if opts.Affected && deps.targetResolver != nil {
		baseRef := opts.BaseRef
		if baseRef == "" {
			baseRef = project.DefaultBranch
		}
		changedLines, clErr := deps.source.ChangedLines(ctx, workspace, baseRef)
		if clErr != nil {
			return pipeline.Result{}, fmt.Errorf("changed lines: %writer", clErr)
		}
		if len(changedLines) > 0 {
			changedFiles := make([]string, 0, len(changedLines))
			for f := range changedLines {
				changedFiles = append(changedFiles, f)
			}
			affected, affErr := deps.targetResolver.FindAffectedTargets(ctx, workspace, changedFiles, project.TargetPattern)
			if affErr != nil {
				return pipeline.Result{}, fmt.Errorf("find affected targets: %writer", affErr)
			}
			scoped := scopeTargetsToPattern(affected, project.TargetPattern, project.ExcludePatterns)
			if len(scoped) > 0 {
				cmdOpts = append(cmdOpts, collectevidence.WithScopedTargets(scoped))
			}
			deps.log.Debug("affected analysis",
				"changed_files", len(changedFiles),
				"affected_targets", len(affected),
				"scoped_targets", len(scoped))
		}
	}

	if len(project.ExcludePatterns) > 0 {
		cmdOpts = append(cmdOpts, collectevidence.WithExcludePatterns(project.ExcludePatterns))
	}
	if len(project.ToolSelection) > 0 {
		cmdOpts = append(cmdOpts, collectevidence.WithToolSelection(project.ToolSelection))
	}

	collectCmd, err := collectevidence.NewCommand(workspace, project.TargetPattern, project.Name, project.DefaultBranch, project.Languages, opts.Quick, opts.Absolute, project.Baseline.ArchIDs, cmdOpts...)
	if err != nil {
		return pipeline.Result{}, err
	}

	if interactive {
		spinner := ui.NewSpinnerWithMessage(writer, "collecting evidence")
		go spinner.Run()
		collected, err := deps.collectEvH.Execute(ctx, collectCmd)
		spinner.Stop()
		if err != nil {
			return pipeline.Result{}, err
		}
		if collected.BuildWarning != "" {
			deps.log.Warn("bazel build had partial failures", "detail", collected.BuildWarning)
		}

		spinner = ui.NewSpinnerWithMessage(writer, "evaluating quality gate")
		go spinner.Run()
		result, err := pipeline.RunProject(ctx, pipeline.Deps{
			Log:          deps.log,
			Submit:       deps.submitH,
			Findings:     deps.findings,
			Coverage:     deps.coverage,
			ServerClient: deps.serverClient,
		}, workspace, collected, tenantID, project.ID, project.Name, commitSHA, branch, startedAt, pipeline.Options{
			Quick:         opts.Quick,
			Absolute:      opts.Absolute,
			RequireSubmit: opts.RequireSubmit,
			PRNumber:      opts.PRNumber,
			PRTitle:       opts.PRTitle,
			PRAuthor:      opts.PRAuthor,
			PRBranch:      opts.PRBranch,
			Gavelspace:    opts.Gavelspace,
			TargetPattern: project.TargetPattern,
			Workspace:     workspace,
		})
		spinner.Stop()
		return result, err
	}

	collected, err := deps.collectEvH.Execute(ctx, collectCmd)
	if err != nil {
		return pipeline.Result{}, err
	}
	if collected.BuildWarning != "" {
		deps.log.Warn("bazel build had partial failures", "detail", collected.BuildWarning)
	}

	return pipeline.RunProject(ctx, pipeline.Deps{
		Log:          deps.log,
		Submit:       deps.submitH,
		Findings:     deps.findings,
		Coverage:     deps.coverage,
		ServerClient: deps.serverClient,
	}, workspace, collected, tenantID, project.ID, project.Name, commitSHA, branch, startedAt, pipeline.Options{
		Quick:         opts.Quick,
		Absolute:      opts.Absolute,
		RequireSubmit: opts.RequireSubmit,
		PRNumber:      opts.PRNumber,
		PRTitle:       opts.PRTitle,
		PRAuthor:      opts.PRAuthor,
		PRBranch:      opts.PRBranch,
		Gavelspace:    opts.Gavelspace,
		TargetPattern: project.TargetPattern,
		Workspace:     workspace,
	})
}
