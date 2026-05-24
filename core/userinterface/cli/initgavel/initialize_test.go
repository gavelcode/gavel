package initgavel

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validGavelYAML(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "gavel.yaml")
	content := `name: test-project
projects:
  - name: api
    pattern: //api/...
    tooling: [go]
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func testCmd(buf *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	return cmd
}

func workspaceResolver(path string) WorkspaceResolver {
	return func() (string, error) { return path, nil }
}

func TestRun_FromConfig(t *testing.T) {
	var buf bytes.Buffer
	workspace := t.TempDir()
	sourceFile := validGavelYAML(t)

	opts := Options{From: sourceFile}
	err := run(testCmd(&buf), opts, workspaceResolver(workspace), &fakeInstaller{}, &fakeCatalog{})

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "INIT")
	assert.Contains(t, output, "CONFIG")
	assert.Contains(t, output, "CREATED")
	assert.Contains(t, output, "BAZEL")
	assert.Contains(t, output, "MODULE")
	assert.Contains(t, output, "test-project")
	assert.Contains(t, output, "SO ORDERED")
}

func TestRun_FromConfigReadError(t *testing.T) {
	var buf bytes.Buffer
	workspace := t.TempDir()

	opts := Options{From: "/nonexistent/gavel.yaml"}
	err := run(testCmd(&buf), opts, workspaceResolver(workspace), &fakeInstaller{}, &fakeCatalog{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "read")
}

func TestRun_WorkspaceError(t *testing.T) {
	var buf bytes.Buffer
	resolver := func() (string, error) { return "", errors.New("no workspace") }

	err := run(testCmd(&buf), Options{From: "x.yaml"}, resolver, &fakeInstaller{}, &fakeCatalog{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no workspace")
}

func TestRun_ModifiedFiles(t *testing.T) {
	var buf bytes.Buffer
	workspace := t.TempDir()
	sourceFile := validGavelYAML(t)

	inst := &fakeInstaller{}
	opts := Options{From: sourceFile}
	err := run(testCmd(&buf), opts, workspaceResolver(workspace), inst, &fakeCatalog{})

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, ".bazelrc")
	assert.Contains(t, output, "MODULE.bazel")
	assert.Contains(t, output, "UPDATED")
	assert.Contains(t, output, "golangci_lint_binary")
}

func TestRun_ExecuteError(t *testing.T) {
	var buf bytes.Buffer
	workspace := t.TempDir()
	sourceFile := validGavelYAML(t)

	inst := &fakeInstaller{installErr: errors.New("install failed")}
	opts := Options{From: sourceFile}
	err := run(testCmd(&buf), opts, workspaceResolver(workspace), inst, &fakeCatalog{})

	require.Error(t, err)
	assert.Contains(t, buf.String(), "FAILED")
}

func TestRun_ConfigDefault(t *testing.T) {
	var buf bytes.Buffer
	workspace := t.TempDir()
	sourceFile := validGavelYAML(t)

	opts := Options{From: sourceFile}
	err := run(testCmd(&buf), opts, workspaceResolver(workspace), &fakeInstaller{}, &fakeCatalog{})

	require.NoError(t, err)
	_, statErr := os.Stat(filepath.Join(workspace, ".gavel/gavel.yaml"))
	assert.NoError(t, statErr)
}

func TestRun_HeaderWriteError(t *testing.T) {
	w := &errWriter{}
	cmd := &cobra.Command{}
	cmd.SetOut(w)

	err := run(cmd, Options{From: "x.yaml"}, workspaceResolver("/tmp"), &fakeInstaller{}, &fakeCatalog{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write error")
}

type errWriter struct{}

func (e *errWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write error")
}

func TestFileStatus_Modified(t *testing.T) {
	modified := map[string]bool{".bazelrc": true}
	assert.Equal(t, "UPDATED", fileStatus(modified, ".bazelrc", "UPDATED"))
}

func TestFileStatus_Unchanged(t *testing.T) {
	modified := map[string]bool{}
	assert.Equal(t, "UNCHANGED", fileStatus(modified, ".bazelrc", "UPDATED"))
}

func TestPrintProjectSummary_FormatsOutput(t *testing.T) {
	var buf bytes.Buffer
	projects := []Project{
		{Name: "api", Pattern: "//api/...", Tooling: []string{"go", "java"}},
		{Name: "web", Pattern: "//web/...", Tooling: []string{"typescript"}},
	}

	err := printProjectSummary(&buf, projects)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "api")
	assert.Contains(t, buf.String(), "//api/...")
	assert.Contains(t, buf.String(), "go, java")
	assert.Contains(t, buf.String(), "web")
}

func TestPrintProjectSummary_WriteError(t *testing.T) {
	projects := []Project{{Name: "api", Pattern: "//api/...", Tooling: []string{"go"}}}

	err := printProjectSummary(&errWriter{}, projects)

	require.Error(t, err)
}

func TestPrintProjectSummary_Empty(t *testing.T) {
	var buf bytes.Buffer

	err := printProjectSummary(&buf, nil)

	require.NoError(t, err)
	assert.Empty(t, buf.String())
}

type countFailWriter struct {
	n   int
	max int
}

func (w *countFailWriter) Write(p []byte) (int, error) {
	w.n++
	if w.n > w.max {
		return 0, errors.New("write error")
	}
	return len(p), nil
}

func TestRun_WriteErrorsAtEachPoint(t *testing.T) {
	for maxWrites := 1; maxWrites <= 15; maxWrites++ {
		t.Run(fmt.Sprintf("failAfter%d", maxWrites), func(t *testing.T) {
			workspace := t.TempDir()
			sourceFile := validGavelYAML(t)
			w := &countFailWriter{max: maxWrites}
			cmd := &cobra.Command{}
			cmd.SetOut(w)
			opts := Options{From: sourceFile}

			err := run(cmd, opts, workspaceResolver(workspace), &fakeInstaller{}, &fakeCatalog{})

			require.Error(t, err)
		})
	}
}

func TestRun_ExecuteErrorWriteFails(t *testing.T) {
	workspace := t.TempDir()
	sourceFile := validGavelYAML(t)
	w := &countFailWriter{max: 3}
	cmd := &cobra.Command{}
	cmd.SetOut(w)
	inst := &fakeInstaller{installErr: errors.New("install failed")}
	opts := Options{From: sourceFile}

	err := run(cmd, opts, workspaceResolver(workspace), inst, &fakeCatalog{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write error")
}

func TestNewCommand_ReturnsCommand(t *testing.T) {
	cmd := NewCommand(
		func() (string, error) { return "/tmp", nil },
		&fakeInstaller{},
		&fakeCatalog{},
	)
	assert.Equal(t, "init", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}
