package watch_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/supporting/analyzetarget"
	"github.com/usegavel/gavel/core/userinterface/cli/watch"
)

type fakeAnalyzer struct {
	mu       sync.Mutex
	findings []analyzetarget.Finding
	err      error
	calls    []string
}

func (f *fakeAnalyzer) Analyze(_ context.Context, _, target string, _ []string) ([]analyzetarget.Finding, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, target)
	return f.findings, f.err
}

type fakeResolver struct {
	mu      sync.Mutex
	targets []string
	err     error
	calls   int
}

func (f *fakeResolver) FindAffectedTargets(_ context.Context, _ string, _ []string, _ string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.targets, f.err
}

func TestWatchEmitsFullEventSequence(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(workspace, "pkg"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "pkg", "main.go"), []byte("package main"), 0o644))

	analyzer := &fakeAnalyzer{findings: []analyzetarget.Finding{{
		Tool:        "golangci-lint",
		RuleID:      "errcheck",
		Severity:    "warning",
		FilePath:    "pkg/main.go",
		Line:        10,
		Message:     "error return not checked",
		Fingerprint: "fp-001",
	}}}
	resolver := &fakeResolver{targets: []string{"//pkg:main"}}
	handler := analyzetarget.NewHandler(analyzer)

	var buf safeBuffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := watch.Options{
		Debounce:  50 * time.Millisecond,
		Languages: []string{"go"},
		Workspace: workspace,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- watch.Run(ctx, &buf, opts, handler, resolver)
	}()

	time.Sleep(200 * time.Millisecond)
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "pkg", "main.go"), []byte("package main\n// changed"), 0o644))

	require.Eventually(t, func() bool {
		return countEvents(&buf, "analysis_done") >= 1
	}, 5*time.Second, 50*time.Millisecond, "timed out waiting for analysis_done event")

	cancel()
	<-errCh

	events := parseEvents(t, &buf)
	require.NotEmpty(t, events)

	assertEventOrder(t, events, []string{"started", "changed", "affected", "analysis_started", "finding", "analysis_done", "stopped"})

	started := findEvent(events, "started")
	require.NotNil(t, started)
	assert.Equal(t, workspace, started.Workspace)

	finding := findEvent(events, "finding")
	require.NotNil(t, finding)
	assert.Equal(t, "golangci-lint", finding.Tool)
	assert.Equal(t, "errcheck", finding.Rule)
	assert.Equal(t, "warning", finding.Severity)
	assert.Equal(t, "pkg/main.go", finding.File)
	assert.Equal(t, 10, finding.Line)
	assert.Equal(t, "fp-001", finding.Fingerprint)

	done := findEvent(events, "analysis_done")
	require.NotNil(t, done)
	assert.Equal(t, "//pkg:main", done.Target)
	assert.Equal(t, 1, done.Findings)
}

func TestWatchEmitsAnalysisFailedOnError(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "file.go"), []byte("package x"), 0o644))

	analyzer := &fakeAnalyzer{err: errors.New("bazel build failed")}
	resolver := &fakeResolver{targets: []string{"//pkg:broken"}}
	handler := analyzetarget.NewHandler(analyzer)

	var buf safeBuffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := watch.Options{
		Debounce:  50 * time.Millisecond,
		Languages: []string{"go"},
		Workspace: workspace,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- watch.Run(ctx, &buf, opts, handler, resolver)
	}()

	time.Sleep(200 * time.Millisecond)
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "file.go"), []byte("package x\n// edit"), 0o644))

	require.Eventually(t, func() bool {
		return countEvents(&buf, "analysis_failed") >= 1
	}, 5*time.Second, 50*time.Millisecond)

	cancel()
	<-errCh

	events := parseEvents(t, &buf)
	failed := findEvent(events, "analysis_failed")
	require.NotNil(t, failed)
	assert.Equal(t, "//pkg:broken", failed.Target)
	assert.Contains(t, failed.Reason, "bazel build failed")
}

func TestWatchNoTargetsSkipsAnalysis(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "readme.md"), []byte("# hi"), 0o644))

	analyzer := &fakeAnalyzer{}
	resolver := &fakeResolver{targets: nil}
	handler := analyzetarget.NewHandler(analyzer)

	var buf safeBuffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := watch.Options{
		Debounce:  50 * time.Millisecond,
		Languages: []string{"go"},
		Workspace: workspace,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- watch.Run(ctx, &buf, opts, handler, resolver)
	}()

	time.Sleep(200 * time.Millisecond)
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "readme.md"), []byte("# edited"), 0o644))

	require.Eventually(t, func() bool {
		return countEvents(&buf, "changed") >= 1
	}, 5*time.Second, 50*time.Millisecond)

	time.Sleep(200 * time.Millisecond)
	cancel()
	<-errCh

	events := parseEvents(t, &buf)
	for _, e := range events {
		assert.NotEqual(t, "analysis_started", e.Event, "should not analyze when no targets affected")
	}

	analyzer.mu.Lock()
	assert.Empty(t, analyzer.calls)
	analyzer.mu.Unlock()
}

type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *safeBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *safeBuffer) snapshot() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]byte, s.buf.Len())
	copy(cp, s.buf.Bytes())
	return cp
}

func countEvents(buf *safeBuffer, eventType string) int {
	data := buf.snapshot()
	count := 0
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		var ev watch.Event
		if json.Unmarshal(scanner.Bytes(), &ev) == nil && ev.Event == eventType {
			count++
		}
	}
	return count
}

func parseEvents(t *testing.T, buf *safeBuffer) []watch.Event {
	t.Helper()
	data := buf.snapshot()
	var events []watch.Event
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var ev watch.Event
		require.NoError(t, json.Unmarshal(line, &ev), "invalid JSONL line: %s", string(line))
		events = append(events, ev)
	}
	return events
}

func findEvent(events []watch.Event, eventType string) *watch.Event {
	for i := range events {
		if events[i].Event == eventType {
			return &events[i]
		}
	}
	return nil
}

func assertEventOrder(t *testing.T, events []watch.Event, expectedOrder []string) {
	t.Helper()
	seen := make([]string, 0, len(events))
	for _, e := range events {
		seen = append(seen, e.Event)
	}

	idx := 0
	for _, eventType := range seen {
		if idx < len(expectedOrder) && eventType == expectedOrder[idx] {
			idx++
		}
	}
	assert.Equal(t, len(expectedOrder), idx,
		"expected event order %v but got sequence %v", expectedOrder, seen)
}
