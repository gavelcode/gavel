package watch

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/supporting/analyzetarget"
)

func TestAnalyzeAndEmitSuccess(t *testing.T) {
	analyzer := &fakeTargetAnalyzer{
		findings: []analyzetarget.Finding{
			{Tool: "golangci-lint", RuleID: "errcheck", Severity: "error", FilePath: "main.go", Line: 42, Message: "unchecked", Fingerprint: "fp1"},
		},
	}
	handler := analyzetarget.NewHandler(analyzer)
	var buf bytes.Buffer
	emit := NewEmitter(&buf)

	analyzeAndEmit(context.Background(), handler, "/ws", "//pkg:lib", []string{"go"}, emit)

	output := buf.String()
	assert.Contains(t, output, `"event":"analysis_started"`)
	assert.Contains(t, output, `"event":"finding"`)
	assert.Contains(t, output, `"tool":"golangci-lint"`)
	assert.Contains(t, output, `"event":"analysis_done"`)
	assert.Contains(t, output, `"findings":1`)
}

func TestAnalyzeAndEmitInvalidCommand(t *testing.T) {
	analyzer := &fakeTargetAnalyzer{}
	handler := analyzetarget.NewHandler(analyzer)
	var buf bytes.Buffer
	emit := NewEmitter(&buf)

	analyzeAndEmit(context.Background(), handler, "", "//pkg:lib", nil, emit)

	output := buf.String()
	assert.Contains(t, output, `"event":"analysis_started"`)
	assert.Contains(t, output, `"event":"analysis_failed"`)
	assert.Contains(t, output, "invalid command")
}

func TestAnalyzeAndEmitHandlerError(t *testing.T) {
	analyzer := &fakeTargetAnalyzer{err: assert.AnError}
	handler := analyzetarget.NewHandler(analyzer)
	var buf bytes.Buffer
	emit := NewEmitter(&buf)

	analyzeAndEmit(context.Background(), handler, "/ws", "//pkg:lib", nil, emit)

	output := buf.String()
	assert.Contains(t, output, `"event":"analysis_started"`)
	assert.Contains(t, output, `"event":"analysis_failed"`)
	assert.Contains(t, output, "analysis failed")
}

func TestDetectWorkspaceSuccess(t *testing.T) {
	setupFakeBazel(t, `echo "/fake/workspace"`)

	workspace, err := detectWorkspace(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "/fake/workspace", workspace)
}

func TestDetectWorkspaceError(t *testing.T) {
	setupFakeBazel(t, "exit 1")

	_, err := detectWorkspace(context.Background())
	require.Error(t, err)
}

func TestDetectBazelVersionSuccess(t *testing.T) {
	dir := setupFakeBazel(t, `echo "Build label: 7.2.0"`)

	version := detectBazelVersion(context.Background(), dir)
	assert.Equal(t, "7.2.0", version)
}

func TestDetectBazelVersionErrorReturnsUnknown(t *testing.T) {
	dir := setupFakeBazel(t, "exit 1")

	version := detectBazelVersion(context.Background(), dir)
	assert.Equal(t, "unknown", version)
}

func TestDetectBazelVersionNoBuildLabelReturnsUnknown(t *testing.T) {
	dir := setupFakeBazel(t, `echo "some other output"`)

	version := detectBazelVersion(context.Background(), dir)
	assert.Equal(t, "unknown", version)
}

func TestNewCommandMetadata(t *testing.T) {
	analyzer := &fakeTargetAnalyzer{}
	handler := analyzetarget.NewHandler(analyzer)
	resolver := &fakeTargetResolver{}

	cmd := NewCommand(handler, resolver)

	assert.Equal(t, "watch", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.Flags().Lookup("debounce"))
	assert.NotNil(t, cmd.Flags().Lookup("languages"))
	assert.NotNil(t, cmd.Flags().Lookup("workspace"))
}

func TestRunEmitsStartedAndStopped(t *testing.T) {
	dir := t.TempDir()
	setupFakeBazel(t, `echo "Build label: 7.0.0"`)

	var buf syncBuffer
	ctx, cancel := context.WithCancel(context.Background())
	analyzer := &fakeTargetAnalyzer{}
	handler := analyzetarget.NewHandler(analyzer)
	resolver := &fakeTargetResolver{}
	opts := Options{Workspace: dir, Debounce: 10 * time.Millisecond}

	errCh := make(chan error, 1)
	go func() { errCh <- Run(ctx, &buf, opts, handler, resolver) }()

	require.Eventually(t, func() bool {
		return strings.Contains(buf.String(), `"event":"started"`)
	}, 5*time.Second, 10*time.Millisecond)
	cancel()
	require.NoError(t, <-errCh)

	output := buf.String()
	assert.Contains(t, output, `"event":"started"`)
	assert.Contains(t, output, `"event":"stopped"`)
	assert.Contains(t, output, dir)
}

func TestRunAutoDetectsWorkspace(t *testing.T) {
	dir := t.TempDir()
	setupFakeBazel(t, `
case "$1" in
  info) echo "`+dir+`" ;;
  version) echo "Build label: 7.0.0" ;;
esac`)

	var buf syncBuffer
	ctx, cancel := context.WithCancel(context.Background())
	analyzer := &fakeTargetAnalyzer{}
	handler := analyzetarget.NewHandler(analyzer)
	resolver := &fakeTargetResolver{}
	opts := Options{Workspace: "", Debounce: 50 * time.Millisecond}

	errCh := make(chan error, 1)
	go func() { errCh <- Run(ctx, &buf, opts, handler, resolver) }()

	require.Eventually(t, func() bool {
		return strings.Contains(buf.String(), `"event":"started"`)
	}, 5*time.Second, 10*time.Millisecond)
	cancel()
	require.NoError(t, <-errCh)

	output := buf.String()
	assert.Contains(t, output, `"event":"started"`)
	assert.Contains(t, output, dir)
}

func TestRunOnChangeAnalyzesTargets(t *testing.T) {
	dir := t.TempDir()
	setupFakeBazel(t, `echo "Build label: 7.0.0"`)

	analyzer := &fakeTargetAnalyzer{
		findings: []analyzetarget.Finding{
			{Tool: "lint", RuleID: "R1", Severity: "warning", FilePath: "a.go", Line: 1, Message: "msg", Fingerprint: "fp"},
		},
	}
	handler := analyzetarget.NewHandler(analyzer)
	resolver := &fakeTargetResolver{targets: []string{"//pkg:lib"}}

	var buf syncBuffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opts := Options{Workspace: dir, Debounce: 10 * time.Millisecond}

	errCh := make(chan error, 1)
	go func() { errCh <- Run(ctx, &buf, opts, handler, resolver) }()

	require.Eventually(t, func() bool {
		return strings.Contains(buf.String(), `"event":"started"`)
	}, 5*time.Second, 10*time.Millisecond)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "trigger.go"), []byte("package x"), 0o644))

	require.Eventually(t, func() bool { return analyzer.called.Load() }, 5*time.Second, 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	cancel()
	<-errCh

	output := buf.String()
	assert.Contains(t, output, `"event":"changed"`)
	assert.Contains(t, output, `"event":"affected"`)
	assert.Contains(t, output, `"event":"analysis_started"`)
	assert.Contains(t, output, `"event":"finding"`)
	assert.Contains(t, output, `"event":"analysis_done"`)
}

func TestRunOnChangeHandlesResolverError(t *testing.T) {
	dir := t.TempDir()
	setupFakeBazel(t, `echo "Build label: 7.0.0"`)

	analyzer := &fakeTargetAnalyzer{}
	handler := analyzetarget.NewHandler(analyzer)
	resolver := &fakeTargetResolver{err: assert.AnError}

	var buf syncBuffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opts := Options{Workspace: dir, Debounce: 10 * time.Millisecond}

	errCh := make(chan error, 1)
	go func() { errCh <- Run(ctx, &buf, opts, handler, resolver) }()

	require.Eventually(t, func() bool {
		return strings.Contains(buf.String(), `"event":"started"`)
	}, 5*time.Second, 10*time.Millisecond)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "trigger.go"), []byte("package x"), 0o644))

	require.Eventually(t, func() bool {
		return strings.Contains(buf.String(), `"event":"changed"`)
	}, 5*time.Second, 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	cancel()
	<-errCh

	output := buf.String()
	assert.Contains(t, output, `"event":"changed"`)
	assert.NotContains(t, output, `"event":"affected"`)
}

func TestRunAutoDetectWorkspaceErrorReturnsError(t *testing.T) {
	setupFakeBazel(t, "exit 1")

	var buf syncBuffer
	analyzer := &fakeTargetAnalyzer{}
	handler := analyzetarget.NewHandler(analyzer)
	resolver := &fakeTargetResolver{}
	opts := Options{Workspace: "", Debounce: 10 * time.Millisecond}

	err := Run(context.Background(), &buf, opts, handler, resolver)
	require.Error(t, err)
}

func TestRunContextCancelledDuringAnalysis(t *testing.T) {
	dir := t.TempDir()
	setupFakeBazel(t, `echo "Build label: 7.0.0"`)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	analyzer := &fakeTargetAnalyzer{
		onAnalyze: func(_ context.Context) {
			cancel()
		},
	}
	handler := analyzetarget.NewHandler(analyzer)
	resolver := &fakeTargetResolver{targets: []string{"//a:lib", "//b:lib", "//c:lib"}}

	var buf syncBuffer
	opts := Options{Workspace: dir, Debounce: 10 * time.Millisecond}

	errCh := make(chan error, 1)
	go func() { errCh <- Run(ctx, &buf, opts, handler, resolver) }()

	require.Eventually(t, func() bool {
		return strings.Contains(buf.String(), `"event":"started"`)
	}, 5*time.Second, 10*time.Millisecond)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "trigger.go"), []byte("package x"), 0o644))

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for Run to exit after context cancellation")
	}
}

func TestNewCommandRunEDelegatesToRun(t *testing.T) {
	dir := t.TempDir()
	setupFakeBazel(t, `echo "Build label: 7.0.0"`)

	analyzer := &fakeTargetAnalyzer{}
	handler := analyzetarget.NewHandler(analyzer)
	resolver := &fakeTargetResolver{}

	cmd := NewCommand(handler, resolver)
	require.NoError(t, cmd.Flags().Set("workspace", dir))
	require.NoError(t, cmd.Flags().Set("debounce", "10ms"))
	cmd.SetOut(&syncBuffer{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	err := cmd.RunE(cmd, nil)
	require.NoError(t, err)
}

func TestRunOnChangeHandlesNoTargets(t *testing.T) {
	dir := t.TempDir()
	setupFakeBazel(t, `echo "Build label: 7.0.0"`)

	analyzer := &fakeTargetAnalyzer{}
	handler := analyzetarget.NewHandler(analyzer)
	resolver := &fakeTargetResolver{targets: nil}

	var buf syncBuffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opts := Options{Workspace: dir, Debounce: 10 * time.Millisecond}

	errCh := make(chan error, 1)
	go func() { errCh <- Run(ctx, &buf, opts, handler, resolver) }()

	require.Eventually(t, func() bool {
		return strings.Contains(buf.String(), `"event":"started"`)
	}, 5*time.Second, 10*time.Millisecond)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "trigger.go"), []byte("package x"), 0o644))

	require.Eventually(t, func() bool {
		return strings.Contains(buf.String(), `"event":"changed"`)
	}, 5*time.Second, 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	cancel()
	<-errCh

	assert.False(t, analyzer.called.Load())
}
