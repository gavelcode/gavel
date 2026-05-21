package runner

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveProjectDir(t *testing.T) {
	tests := []struct {
		workspace string
		pattern   string
		want      string
	}{
		{"/ws", "//apps/web/...", "/ws/apps/web"},
		{"/ws", "//core/...", "/ws/core"},
		{"/ws", "//...", "/ws"},
	}
	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got := resolveProjectDir(tt.workspace, tt.pattern)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHasPackageJSON_Missing(t *testing.T) {
	assert.False(t, hasPackageJSON(t.TempDir()))
}

func TestRunJSCoverage_NoPackageJSON(t *testing.T) {
	workspace := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	data, err := runJSCoverage(t.Context(), &fakeRunner{}, workspace, "//apps/web/...", logger)

	require.NoError(t, err)
	assert.Nil(t, data)
}

const fakeNpxPath = "/usr/bin/npx"

func TestRunJSCoverage_NpxNotFound(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "package.json"), []byte("{}"), 0o644))
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	original := lookPathFunc
	lookPathFunc = func(string) (string, error) { return "", fmt.Errorf("not found") }
	t.Cleanup(func() { lookPathFunc = original })

	_, err := runJSCoverage(t.Context(), &fakeRunner{}, workspace, "//...", logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "npx not found")
}

func TestRunJSCoverage_Success(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "package.json"), []byte("{}"), 0o644))
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	original := lookPathFunc
	lookPathFunc = func(string) (string, error) { return fakeNpxPath, nil }
	t.Cleanup(func() { lookPathFunc = original })

	fake := &fakeRunner{
		results: []fakeResult{{Stdout: []byte("vitest ok\n")}},
		runHook: func(_ string, args []string) {
			for _, arg := range args {
				if coverageDir, ok := strings.CutPrefix(arg, "--coverage.reportsDirectory="); ok {
					_ = os.WriteFile(filepath.Join(coverageDir, "lcov.info"), []byte("SF:app.ts\nDA:1,1\nend_of_record\n"), 0o644)
				}
			}
		},
	}

	data, err := runJSCoverage(t.Context(), fake, workspace, "//...", logger)

	require.NoError(t, err)
	require.NotNil(t, data)
	assert.Contains(t, string(data), "SF:app.ts")
}

func TestRunJSCoverage_VitestOtherError(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "package.json"), []byte("{}"), 0o644))
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	original := lookPathFunc
	lookPathFunc = func(string) (string, error) { return fakeNpxPath, nil }
	t.Cleanup(func() { lookPathFunc = original })

	fake := &fakeRunner{
		results: []fakeResult{{Stderr: []byte("timeout"), Err: fmt.Errorf("context deadline exceeded")}},
	}

	_, err := runJSCoverage(t.Context(), fake, workspace, "//...", logger)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "vitest coverage")
}

func TestRunJSCoverage_NoCoverageFile(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "package.json"), []byte("{}"), 0o644))
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	original := lookPathFunc
	lookPathFunc = func(string) (string, error) { return fakeNpxPath, nil }
	t.Cleanup(func() { lookPathFunc = original })

	fake := &fakeRunner{results: []fakeResult{{Stdout: []byte("ok\n")}}}

	data, err := runJSCoverage(t.Context(), fake, workspace, "//...", logger)

	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestRunJSCoverage_EmptyCoverageFile(t *testing.T) {
	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "package.json"), []byte("{}"), 0o644))
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	original := lookPathFunc
	lookPathFunc = func(string) (string, error) { return fakeNpxPath, nil }
	t.Cleanup(func() { lookPathFunc = original })

	fake := &fakeRunner{
		results: []fakeResult{{Stdout: []byte("ok\n")}},
		runHook: func(_ string, args []string) {
			for _, arg := range args {
				if coverageDir, ok := strings.CutPrefix(arg, "--coverage.reportsDirectory="); ok {
					_ = os.WriteFile(filepath.Join(coverageDir, "lcov.info"), nil, 0o644)
				}
			}
		},
	}

	data, err := runJSCoverage(t.Context(), fake, workspace, "//...", logger)

	require.NoError(t, err)
	assert.Nil(t, data)
}
