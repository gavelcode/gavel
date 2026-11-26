package runner_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/runner"
)

func TestExtractToolName_ValidSARIF(t *testing.T) {
	data := []byte(`{"runs":[{"tool":{"driver":{"name":"SpotBugs"}},"results":[]}]}`)

	got := runner.ExtractToolName(data, "fallback.report")

	assert.Equal(t, "SpotBugs", got)
}

func TestExtractToolName_InvalidJSON(t *testing.T) {
	got := runner.ExtractToolName([]byte("not json"), "myfile.report")

	assert.Equal(t, "myfile.report", got)
}

func TestExtractToolName_EmptyToolName(t *testing.T) {
	data := []byte(`{"runs":[{"tool":{"driver":{"name":""}},"results":[]}]}`)

	got := runner.ExtractToolName(data, "fallback.report")

	assert.Equal(t, "fallback.report", got)
}

func TestExtractToolName_NoRuns(t *testing.T) {
	data := []byte(`{"runs":[]}`)

	got := runner.ExtractToolName(data, "fallback.report")

	assert.Equal(t, "fallback.report", got)
}
