package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bepHook plays bazel: it writes the given files and a build event stream
// that reports exactly `reported` as the aspect's gavel_submissions outputs.
func bepHook(t *testing.T, dir string, reported []string, nested bool) func(string, []string) {
	t.Helper()
	return func(_ string, args []string) {
		var bepPath string
		for _, arg := range args {
			if value, ok := strings.CutPrefix(arg, "--build_event_json_file="); ok {
				bepPath = value
			}
		}
		require.NotEmpty(t, bepPath, "bridge runs must ask bazel for the build event stream")
		var lines []string
		files := make([]map[string]any, 0, len(reported))
		for _, name := range reported {
			files = append(files, map[string]any{"name": name, "uri": "file://" + filepath.Join(dir, name)})
		}
		leaf := map[string]any{"id": map[string]any{"namedSet": map[string]any{"id": "0"}}, "namedSetOfFiles": map[string]any{"files": files}}
		sets := []map[string]any{{"id": "0"}}
		lines = append(lines, mustJSON(t, leaf))
		if nested {
			parent := map[string]any{"id": map[string]any{"namedSet": map[string]any{"id": "1"}}, "namedSetOfFiles": map[string]any{"fileSets": []map[string]any{{"id": "0"}}}}
			lines = append(lines, mustJSON(t, parent))
			sets = []map[string]any{{"id": "1"}}
		}
		completed := map[string]any{
			"id": map[string]any{"targetCompleted": map[string]any{"label": "//app:scripts", "aspect": "//tools/lint:todo.bzl%todo_gavel"}},
			"completed": map[string]any{"success": true, "outputGroup": []map[string]any{
				{"name": "gavel_submissions", "fileSets": sets},
				{"name": "other_group", "fileSets": []map[string]any{{"id": "9"}}},
			}},
		}
		lines = append(lines, mustJSON(t, completed))
		require.NoError(t, os.WriteFile(bepPath, []byte(strings.Join(lines, "\n")+"\n"), 0o600))
	}
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	return string(data)
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
}

func bridgeConfig() BridgeConfig {
	return BridgeConfig{
		Workspace:   "/ws",
		Targets:     []string{"//app/...", "-//app/gen/..."},
		Aspect:      "//tools/lint:todo.bzl%todo_gavel",
		SARIFSuffix: ".todo.sarif",
		BuildFlags:  []string{"--define=X=1"},
	}
}

func TestRunBridgeBuildsTheAspectWithItsOwnOutputGroupAndFlags(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app/scripts.todo.sarif", `{"runs":[]}`)
	fake := &fakeRunner{results: []fakeResult{{}}, runHook: bepHook(t, dir, []string{"app/scripts.todo.sarif"}, false)}
	config := bridgeConfig()
	config.BazelConfigs = []string{"ci"}

	_, err := runBridge(t.Context(), fake, config)

	require.NoError(t, err)
	require.Len(t, fake.calls, 1)
	args := fake.calls[0].Args
	assert.Equal(t, "/ws", fake.calls[0].Dir)
	assert.Equal(t, "build", args[0])
	assert.Contains(t, args, "--config=ci")
	assert.Contains(t, args, "--aspects=//tools/lint:todo.bzl%todo_gavel")
	assert.Contains(t, args, "--output_groups=gavel_submissions")
	assert.Contains(t, args, "--define=X=1")
	assert.Contains(t, args, "--remote_download_toplevel")
	separator := indexOf(args, "--")
	require.NotEqual(t, -1, separator)
	assert.Equal(t, []string{"//app/...", "-//app/gen/..."}, args[separator+1:])
}

func TestRunBridgeCollectsOnlyFilesThisBuildReported(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app/scripts.todo.sarif", `{"fresh":true}`)
	writeFile(t, dir, "old/stale.todo.sarif", `{"stale":true}`)
	writeFile(t, dir, "app/other.lint.sarif", `{"other":true}`)
	fake := &fakeRunner{results: []fakeResult{{}}, runHook: bepHook(t, dir, []string{"app/scripts.todo.sarif", "app/other.lint.sarif"}, false)}

	result, err := runBridge(t.Context(), fake, bridgeConfig())

	require.NoError(t, err)
	require.Len(t, result.Files, 1, "stale files left in bazel-bin and other suffixes are never read")
	assert.Equal(t, "app/scripts.todo.sarif", result.Files[0].Name)
	assert.JSONEq(t, `{"fresh":true}`, string(result.Files[0].Data))
}

func TestRunBridgeFollowsNestedFileSets(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app/scripts.todo.sarif", `{"runs":[]}`)
	fake := &fakeRunner{results: []fakeResult{{}}, runHook: bepHook(t, dir, []string{"app/scripts.todo.sarif"}, true)}

	result, err := runBridge(t.Context(), fake, bridgeConfig())

	require.NoError(t, err)
	assert.Len(t, result.Files, 1)
}

func TestRunBridgeFailsWhenTheBuildFails(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app/scripts.todo.sarif", `{"runs":[]}`)
	fake := &fakeRunner{
		results: []fakeResult{{Stderr: []byte("ERROR: aspect exploded\nline2"), Err: errors.New("exit status 1")}},
		runHook: bepHook(t, dir, []string{"app/scripts.todo.sarif"}, false),
	}

	result, err := runBridge(t.Context(), fake, bridgeConfig())

	require.Error(t, err)
	assert.Nil(t, result, "a failed build is never half-trusted")
	assert.Contains(t, err.Error(), "bazel build failed")
	assert.Contains(t, err.Error(), "aspect exploded")
}

func TestRunBridgeFailsWhenNoSARIFWithTheSuffixWasProduced(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app/other.lint.sarif", `{}`)
	fake := &fakeRunner{results: []fakeResult{{}}, runHook: bepHook(t, dir, []string{"app/other.lint.sarif"}, false)}

	_, err := runBridge(t.Context(), fake, bridgeConfig())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no SARIF files")
	assert.Contains(t, err.Error(), ".todo.sarif")
	assert.Contains(t, err.Error(), "gavel_submissions")
}

func TestRunBridgeBoundsTheBuildOutputInTheFailureReason(t *testing.T) {
	var lines []string
	for i := 0; i < 100; i++ {
		lines = append(lines, fmt.Sprintf("noise %d", i))
	}
	fake := &fakeRunner{results: []fakeResult{{Stderr: []byte(strings.Join(lines, "\n")), Err: errors.New("exit status 1")}}, runHook: bepHook(t, t.TempDir(), nil, false)}

	_, err := runBridge(t.Context(), fake, bridgeConfig())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "noise 99")
	assert.NotContains(t, err.Error(), "noise 0\n")
}
