package watch

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/usegavel/gavel/core/application/supporting/analyzetarget"
)

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

type fakeTargetAnalyzer struct {
	findings  []analyzetarget.Finding
	err       error
	called    atomic.Bool
	onAnalyze func(ctx context.Context)
}

func (f *fakeTargetAnalyzer) Analyze(ctx context.Context, _, _ string, _ []string) ([]analyzetarget.Finding, error) {
	f.called.Store(true)
	if f.onAnalyze != nil {
		f.onAnalyze(ctx)
	}
	return f.findings, f.err
}

type fakeTargetResolver struct {
	targets []string
	err     error
}

func (f *fakeTargetResolver) FindAffectedTargets(_ context.Context, _ string, _ []string, _ string) ([]string, error) {
	return f.targets, f.err
}

func setupFakeBazel(t *testing.T, script string) string {
	t.Helper()
	dir := t.TempDir()
	fakePath := filepath.Join(dir, "bazel")
	if err := os.WriteFile(fakePath, fmt.Appendf(nil, "#!/bin/sh\n%s", script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	return dir
}
