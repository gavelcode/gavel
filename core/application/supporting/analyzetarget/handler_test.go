package analyzetarget_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/supporting/analyzetarget"
)

type fakeAnalyzer struct {
	findings []analyzetarget.Finding
	err      error
}

func (f *fakeAnalyzer) Analyze(_ context.Context, _, _ string, _ []string) ([]analyzetarget.Finding, error) {
	return f.findings, f.err
}

func TestExecute_ReturnsFindings(t *testing.T) {
	analyzer := &fakeAnalyzer{
		findings: []analyzetarget.Finding{
			{Tool: "golangci-lint", RuleID: "unused", Severity: "warning", FilePath: "main.go", Line: 10, Message: "unused var"},
		},
	}
	h := analyzetarget.NewHandler(analyzer)

	cmd, err := analyzetarget.NewCommand("/ws", "//pkg:lib", []string{"go"})
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, "//pkg:lib", result.Target)
	assert.Len(t, result.Findings, 1)
	assert.Equal(t, "unused", result.Findings[0].RuleID)
	assert.True(t, result.Duration > 0)
}

func TestExecute_AnalyzerError(t *testing.T) {
	analyzer := &fakeAnalyzer{err: fmt.Errorf("bazel build failed")}
	h := analyzetarget.NewHandler(analyzer)

	cmd, err := analyzetarget.NewCommand("/ws", "//pkg:lib", []string{"go"})
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), cmd)

	assert.Error(t, err)
}

func TestNewCommand_EmptyTarget(t *testing.T) {
	_, err := analyzetarget.NewCommand("/ws", "", []string{"go"})
	assert.ErrorIs(t, err, analyzetarget.ErrInvalidCommand)
}

func TestNewCommand_EmptyWorkspace(t *testing.T) {
	_, err := analyzetarget.NewCommand("", "//pkg:lib", []string{"go"})
	assert.ErrorIs(t, err, analyzetarget.ErrInvalidCommand)
}

func TestNewHandler_NilAnalyzerPanics(t *testing.T) {
	assert.Panics(t, func() {
		analyzetarget.NewHandler(nil)
	})
}
