package qualitygate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

func TestMinNewCodeCoverageShouldPassAboveThreshold(t *testing.T) {
	strategy, err := qualitygate.NewMinNewCodeCoverage(80)
	require.NoError(t, err)

	content, err := coverage.NewPatchContent(9, 10)
	require.NoError(t, err)

	outcome := strategy.Evaluate(content)

	assert.True(t, outcome.Passed())
}

func TestMinNewCodeCoverageShouldPassAtExactThreshold(t *testing.T) {
	strategy, err := qualitygate.NewMinNewCodeCoverage(80)
	require.NoError(t, err)

	content, err := coverage.NewPatchContent(8, 10)
	require.NoError(t, err)

	outcome := strategy.Evaluate(content)

	assert.True(t, outcome.Passed())
}

func TestMinNewCodeCoverageShouldFailBelowThreshold(t *testing.T) {
	strategy, err := qualitygate.NewMinNewCodeCoverage(80)
	require.NoError(t, err)

	content, err := coverage.NewPatchContent(7, 10)
	require.NoError(t, err)

	outcome := strategy.Evaluate(content)

	assert.False(t, outcome.Passed())
	assert.Contains(t, outcome.Detail(), "70.0%")
	assert.Contains(t, outcome.Detail(), "80.0%")
}

func TestMinNewCodeCoverageShouldPassWithZeroCoverableLines(t *testing.T) {
	strategy, err := qualitygate.NewMinNewCodeCoverage(80)
	require.NoError(t, err)

	content, err := coverage.NewPatchContent(0, 0)
	require.NoError(t, err)

	outcome := strategy.Evaluate(content)

	assert.True(t, outcome.Passed())
}

func TestMinNewCodeCoverageShouldPassForNonMatchingContent(t *testing.T) {
	strategy, err := qualitygate.NewMinNewCodeCoverage(80)
	require.NoError(t, err)

	content, err := coverage.NewContent(100, 80, nil)
	require.NoError(t, err)

	outcome := strategy.Evaluate(content)

	assert.True(t, outcome.Passed())
}

func TestMinNewCodeCoverageShouldRejectNegativeMin(t *testing.T) {
	_, err := qualitygate.NewMinNewCodeCoverage(-1)
	assert.Error(t, err)
}

func TestMinNewCodeCoverageShouldRejectMinAboveHundred(t *testing.T) {
	_, err := qualitygate.NewMinNewCodeCoverage(101)
	assert.Error(t, err)
}
