package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/usegavel/gavel/apps/cli/internal/app"
	"github.com/usegavel/gavel/core/application/casefile/classify"
	"github.com/usegavel/gavel/core/application/casefile/classifyarch"
	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	"github.com/usegavel/gavel/core/application/casefile/createcasefile"
	"github.com/usegavel/gavel/core/application/casefile/finalize"
	ingestcov "github.com/usegavel/gavel/core/application/casefile/ingestcoverage"
	"github.com/usegavel/gavel/core/application/casefile/ingestevidence"
	ingestfind "github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	"github.com/usegavel/gavel/core/application/casefile/ingestncc"
	corejudge "github.com/usegavel/gavel/core/application/casefile/judge"
	"github.com/usegavel/gavel/core/application/casefile/submit"
	"github.com/usegavel/gavel/core/application/gavelspace/loadgavelspace"
	"github.com/usegavel/gavel/core/application/supporting/analyzetarget"
	"github.com/usegavel/gavel/core/infrastructure/casefile/lcov"
	casefilememory "github.com/usegavel/gavel/core/infrastructure/casefile/memory"
	"github.com/usegavel/gavel/core/infrastructure/casefile/sarif"
	"github.com/usegavel/gavel/core/infrastructure/gavelspace/gavelconfig"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/catalog"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/collector"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/collector/composite"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/installer"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/runner"
	"github.com/usegavel/gavel/core/infrastructure/platform/git"
	"github.com/usegavel/gavel/core/infrastructure/project/archconfig"
	projectmemory "github.com/usegavel/gavel/core/infrastructure/project/memory"
	"github.com/usegavel/gavel/core/userinterface/cli/judge"
)

func main() {
	catalog.SetModulePrefix("@gavel_tools")

	workspace, wsErr := bazel.WorkspaceDir()

	var projectRepo *projectmemory.ProjectRepository
	if wsErr == nil {
		projectRepo = projectmemory.NewProjectRepositoryWithBaseline(projectmemory.NewBaselineStore(workspace))
	} else {
		projectRepo = projectmemory.NewProjectRepository()
	}
	caseFileRepo := casefilememory.NewCaseFileRepository()

	var sarifOpts []sarif.ParserOption
	if wsErr == nil {
		sarifOpts = append(sarifOpts, sarif.WithSourceReader(sarif.NewFileSourceReader(workspace)))
	}
	sarifParser := sarif.NewParser(sarifOpts...)
	lcovParser := lcov.NewParser()

	findingsHandler := ingestfind.NewHandler(map[string]ingestfind.Parser{"sarif": sarifParser})
	coverageHandler := ingestcov.NewHandler(map[string]ingestcov.Parser{"lcov": lcovParser})

	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelWarn)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))

	classifyHandler := classify.NewHandler(caseFileRepo)
	judgeHandler := corejudge.NewHandler(caseFileRepo, projectRepo)
	createCFHandler := createcasefile.NewHandler(caseFileRepo, projectRepo)
	ingestEvHandler := ingestevidence.NewHandler(caseFileRepo)
	finalizeHandler := finalize.NewHandler(caseFileRepo, projectRepo, classifyHandler, judgeHandler, nil, finalize.WithLogger(logger))
	nccHandler := ingestncc.NewHandler(lcovParser)
	classifyArchHandler := classifyarch.NewHandler()

	bazelRunner := collector.NewBazelRunner()
	findingsCollector := collector.NewBazelFindingsCollector(bazelRunner, findingsHandler)
	bazelCoverage := collector.NewBazelCoverageCollector(bazelRunner)
	vitestCoverage := collector.NewVitestCoverageCollector(logger)
	coverageCollector := composite.NewCoverageCollector(bazelCoverage, vitestCoverage)
	archCollector := collector.NewBazelArchitectureCollector(bazelRunner)
	sourceContext := git.NewSourceContext()

	collectEvHandler := collectevidence.NewHandler(
		findingsCollector, coverageCollector, archCollector,
		findingsHandler, coverageHandler,
		classifyArchHandler, nccHandler,
		collectevidence.WithChangedLinesSource(sourceContext),
		collectevidence.WithPerLineParser(lcovParser),
	)
	inst := installer.NewInstaller()
	submitHandler := submit.NewHandler(createCFHandler, ingestEvHandler, finalizeHandler)
	loadWsHandler := loadgavelspace.NewHandler(gavelconfig.NewWorkspaceFinder(),
		loadgavelspace.WithArchPolicyLoader(archconfig.NewPolicyLoader()),
		loadgavelspace.WithProjectSaver(projectRepo),
		loadgavelspace.WithLogger(logger),
	)
	execRunner := runner.NewExecRunner()
	targetResolver := runner.NewBazelTargetResolver(execRunner)
	targetAnalyzer := runner.NewBazelTargetAnalyzer(execRunner)
	analyzeHandler := analyzetarget.NewHandler(targetAnalyzer)

	deps := app.Deps{
		WorkspaceResolver:   bazel.WorkspaceDir,
		Logger:              logger,
		LogLevel:            logLevel,
		FindingsHandler:     findingsHandler,
		CoverageHandler:     coverageHandler,
		SubmitHandler:       submitHandler,
		CollectEvHandler:    collectEvHandler,
		LoadWsHandler:       loadWsHandler,
		ProjectRepo:         projectRepo,
		FPSeeder:            caseFileRepo,
		SourceContext:       sourceContext,
		TargetQuery:         runner.NewBazelTargetQuery(execRunner),
		Verifier:            inst,
		ConfigInstaller:     inst,
		ToolCatalog:         inst,
		AnalyzeHandler:      analyzeHandler,
		JudgeTargetResolver: targetResolver,
		WatchTargetResolver: targetResolver,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	root := app.NewRootCommand(deps)
	root.SetContext(ctx)
	if err := root.Execute(); err != nil {
		if !errors.Is(err, judge.ErrVerdictFail) {
			root.PrintErrln(err)
		}
		os.Exit(1)
	}
}
