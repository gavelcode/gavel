package watch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/usegavel/gavel/core/application/supporting/analyzetarget"
)

const (
	detectWorkspaceTimeout = 30 * time.Second
	// Version detection is best-effort metadata for the started event; cap it
	// low so a cold/contended bazel never delays the watcher from registering.
	detectVersionTimeout = 5 * time.Second
)

func NewCommand(handler *analyzetarget.Handler, resolver analyzetarget.TargetResolver) *cobra.Command {
	opts := Options{}
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch workspace and emit JSONL analysis events",
		Long: `Watch reacts to file changes by running bazel lint aspects on affected
targets and emits one JSON event per line to stdout. Designed to be consumed
by tooling — LLM monitor modes, CI pipelines, IDE bridges.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return Run(cmd.Context(), os.Stdout, opts, handler, resolver)
		},
	}
	RegisterFlags(cmd, &opts)
	return cmd
}

func Run(ctx context.Context, stdout io.Writer, opts Options, handler *analyzetarget.Handler, resolver analyzetarget.TargetResolver) error {
	if opts.Workspace == "" {
		ws, err := detectWorkspace(ctx)
		if err != nil {
			return err
		}
		opts.Workspace = ws
	}
	emit := NewEmitter(stdout)

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	onChange := func(ctx context.Context, files []string) {
		_ = emit.Changed(files)
		targets, err := resolver.FindAffectedTargets(ctx, opts.Workspace, files, "")
		if err != nil {
			log.Warn("find affected targets", "err", err)
			return
		}
		if len(targets) == 0 {
			return
		}
		_ = emit.Affected(targets)
		for _, target := range targets {
			if ctx.Err() != nil {
				return
			}
			analyzeAndEmit(ctx, handler, opts.Workspace, target, opts.Languages, emit)
		}
	}

	watcher, err := NewFileWatcher(opts.Workspace, opts.Debounce, onChange, log)
	if err != nil {
		return err
	}

	watcher.onStart = func() {
		_ = emit.Started(opts.Workspace, detectBazelVersion(ctx, opts.Workspace))
	}
	defer func() { _ = emit.Stopped("signal") }()
	return watcher.Run(ctx)
}

func analyzeAndEmit(ctx context.Context, handler *analyzetarget.Handler, workspace, target string, languages []string, emit *Emitter) {
	_ = emit.AnalysisStarted(target)

	cmd, err := analyzetarget.NewCommand(workspace, target, languages)
	if err != nil {
		_ = emit.AnalysisFailed(target, fmt.Sprintf("invalid command: %v", err))
		return
	}

	result, err := handler.Execute(ctx, cmd)
	if err != nil {
		_ = emit.AnalysisFailed(target, fmt.Sprintf("analysis failed: %v", err))
		return
	}

	for _, f := range result.Findings {
		_ = emit.Finding(target, f.Tool, f.RuleID, f.Severity, f.FilePath, f.Line, f.Message, f.Fingerprint)
	}
	_ = emit.AnalysisDone(target, len(result.Findings), result.Duration)
}

func detectWorkspace(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, detectWorkspaceTimeout)
	defer cancel()

	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, "bazel", "info", "workspace")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout.String()), nil
}

func detectBazelVersion(ctx context.Context, workspace string) string {
	ctx, cancel := context.WithTimeout(ctx, detectVersionTimeout)
	defer cancel()

	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, "bazel", "version")
	cmd.Dir = workspace
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(stdout.String(), "\n") {
		if strings.HasPrefix(line, "Build label:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Build label:"))
		}
	}
	return "unknown"
}
