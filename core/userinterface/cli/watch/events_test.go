package watch

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmitterProducesLineDelimitedJSON(t *testing.T) {
	var buf bytes.Buffer
	emit := NewEmitter(&buf)
	emit.now = fixedClock("2026-01-01T00:00:00Z")

	require.NoError(t, emit.Started("/ws", "7.2.0"))
	require.NoError(t, emit.Changed([]string{"pkg/foo.go"}))
	require.NoError(t, emit.Affected([]string{"//pkg/a:lib"}))
	require.NoError(t, emit.AnalysisStarted("//pkg/a:lib"))
	require.NoError(t, emit.Finding("//pkg/a:lib", "golangci-lint", "errcheck", "error", "pkg/a/foo.go", 42, "msg", "abc"))
	require.NoError(t, emit.AnalysisDone("//pkg/a:lib", 1, 1340*time.Millisecond))
	require.NoError(t, emit.AnalysisFailed("//pkg/b:lib", "bazel build failed"))
	require.NoError(t, emit.Stopped("signal"))

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 8)

	got := decode(t, lines[0])
	assert.Equal(t, "started", got["event"])
	assert.Equal(t, "/ws", got["workspace"])
	assert.Equal(t, "7.2.0", got["bazel_version"])
	assert.Equal(t, "2026-01-01T00:00:00Z", got["ts"])

	got = decode(t, lines[1])
	assert.Equal(t, "changed", got["event"])
	assert.Equal(t, []any{"pkg/foo.go"}, got["files"])

	got = decode(t, lines[2])
	assert.Equal(t, "affected", got["event"])
	assert.Equal(t, []any{"//pkg/a:lib"}, got["targets"])

	got = decode(t, lines[3])
	assert.Equal(t, "analysis_started", got["event"])
	assert.Equal(t, "//pkg/a:lib", got["target"])

	got = decode(t, lines[4])
	assert.Equal(t, "finding", got["event"])
	assert.Equal(t, "golangci-lint", got["tool"])
	assert.Equal(t, "errcheck", got["rule"])
	assert.Equal(t, "error", got["severity"])
	assert.Equal(t, "pkg/a/foo.go", got["file"])
	assert.Equal(t, float64(42), got["line"])
	assert.Equal(t, "abc", got["fingerprint"])

	got = decode(t, lines[5])
	assert.Equal(t, "analysis_done", got["event"])
	assert.Equal(t, float64(1), got["findings"])
	assert.Equal(t, float64(1340), got["duration_ms"])

	got = decode(t, lines[6])
	assert.Equal(t, "analysis_failed", got["event"])
	assert.Equal(t, "bazel build failed", got["reason"])

	got = decode(t, lines[7])
	assert.Equal(t, "stopped", got["event"])
}

func TestEmitterIsConcurrencySafe(t *testing.T) {
	var buf bytes.Buffer
	emit := NewEmitter(&buf)

	done := make(chan struct{}, 4)
	for i := 0; i < 4; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 50; j++ {
				_ = emit.AnalysisStarted("//x")
			}
		}()
	}
	for i := 0; i < 4; i++ {
		<-done
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 200)
	for _, line := range lines {
		got := decode(t, line)
		assert.Equal(t, "analysis_started", got["event"])
	}
}

func fixedClock(rfc3339 string) func() time.Time {
	t, _ := time.Parse(time.RFC3339, rfc3339)
	return func() time.Time { return t }
}

func decode(t *testing.T, line string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(line), &m))
	return m
}
