//go:build e2e

package e2e_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate_PassesOnConfiguredWorkspace(t *testing.T) {
	t.Cleanup(func() { cleanWorkspace(t) })

	result := runGavel(t, "validate")
	assert.Equal(t, 0, result.ExitCode,
		"validate should pass on a configured workspace\nstdout: %s\nstderr: %s", result.Stdout, result.Stderr)
	assert.True(t, strings.Contains(result.Stdout, "VALID") || strings.Contains(result.Stdout, "valid"),
		"stdout should indicate valid structure, got:\n%s", result.Stdout)
}
