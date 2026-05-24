//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrends_RequiresServer(t *testing.T) {
	cleanWorkspace(t)

	result := runGavel(t, "trends", "--project=api-gateway")
	assert.NotEqual(t, 0, result.ExitCode)
	assert.Contains(t, result.Stderr+result.Stdout, "server")
}

func TestTrends_RequiresProject(t *testing.T) {
	sf := startServer(t)

	result := runGavel(t, "trends", "--server="+sf.URL, "--token="+sf.Token)
	assert.NotEqual(t, 0, result.ExitCode)
	assert.Contains(t, result.Stderr+result.Stdout, "project")
}

func TestTrends_FailsForUnknownProject(t *testing.T) {
	sf := startServer(t)

	result := runGavel(t, "trends", "--project=nonexistent",
		"--server="+sf.URL, "--token="+sf.Token)
	assert.NotEqual(t, 0, result.ExitCode)
}
