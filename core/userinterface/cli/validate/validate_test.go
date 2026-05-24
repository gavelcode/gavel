package validate_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/cli/validate"
)

func stubWorkspaceResolver() (string, error) { return "/tmp/workspace", nil }

func failingWorkspaceResolver() (string, error) { return "", errors.New("no workspace") }

func setupWorkspace(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "MODULE.bazel"), []byte(""), 0o644))
	t.Setenv("BUILD_WORKSPACE_DIRECTORY", dir)
}

type stubVerifier struct {
	issues []string
	err    error
}

func (s stubVerifier) VerifyStructure(_ string) ([]string, error) {
	return s.issues, s.err
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

type failOnNthWriter struct {
	n       int
	written int
}

func (w *failOnNthWriter) Write(p []byte) (int, error) {
	w.written++
	if w.written >= w.n {
		return 0, errors.New("write failed")
	}
	return len(p), nil
}

func TestValidateShowsValidWhenNoIssues(t *testing.T) {
	setupWorkspace(t)
	cmd := validate.NewCommand(stubWorkspaceResolver, stubVerifier{})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "VALID")
	assert.Contains(t, buf.String(), "structure is valid")
}

func TestValidateShowsInvalidWithIssues(t *testing.T) {
	setupWorkspace(t)
	verifier := stubVerifier{issues: []string{"gavel.bazelrc not found", ".bazelrc missing include"}}
	cmd := validate.NewCommand(stubWorkspaceResolver, verifier)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, buf.String(), "INVALID")
	assert.Contains(t, buf.String(), "gavel.bazelrc not found")
	assert.Contains(t, buf.String(), ".bazelrc missing include")
}

func TestValidateReturnsVerifierError(t *testing.T) {
	setupWorkspace(t)
	verifier := stubVerifier{err: errors.New("disk read failed")}
	cmd := validate.NewCommand(stubWorkspaceResolver, verifier)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "verify structure")
}

func TestValidateReturnsWorkspaceResolverError(t *testing.T) {
	cmd := validate.NewCommand(failingWorkspaceResolver, stubVerifier{})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no workspace")
}

func TestValidateReturnsWriteErrorOnValid(t *testing.T) {
	cmd := validate.NewCommand(stubWorkspaceResolver, stubVerifier{})
	cmd.SetOut(failWriter{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}

func TestValidateReturnsWriteErrorOnInvalid(t *testing.T) {
	verifier := stubVerifier{issues: []string{"missing file"}}
	cmd := validate.NewCommand(stubWorkspaceResolver, verifier)
	cmd.SetOut(failWriter{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}

func TestValidateReturnsWriteErrorOnIssueItem(t *testing.T) {
	verifier := stubVerifier{issues: []string{"missing file"}}
	cmd := validate.NewCommand(stubWorkspaceResolver, verifier)
	cmd.SetOut(&failOnNthWriter{n: 2})
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}
