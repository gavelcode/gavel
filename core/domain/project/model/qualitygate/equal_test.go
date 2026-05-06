package qualitygate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

func TestCountBySeverityEqual(t *testing.T) {
	lhs, err := qualitygate.NewCountBySeverity(1, 2, 3)
	require.NoError(t, err)
	rhs, err := qualitygate.NewCountBySeverity(1, 2, 3)
	require.NoError(t, err)
	alt, err := qualitygate.NewCountBySeverity(1, 2, 4)
	require.NoError(t, err)

	assert.True(t, lhs.Equal(rhs))
	assert.True(t, rhs.Equal(lhs))
	assert.False(t, lhs.Equal(alt))
	assert.False(t, lhs.Equal(qualitygate.NewZeroTolerance()))
}

func TestZeroToleranceEqual(t *testing.T) {
	lhs := qualitygate.NewZeroTolerance()
	rhs := qualitygate.NewZeroTolerance()
	alt, err := qualitygate.NewCountBySeverity(0, 0, 0)
	require.NoError(t, err)

	assert.True(t, lhs.Equal(rhs))
	assert.False(t, lhs.Equal(alt))
}

func TestMinPercentageEqual(t *testing.T) {
	lhs, err := qualitygate.NewMinPercentage(80)
	require.NoError(t, err)
	rhs, err := qualitygate.NewMinPercentage(80)
	require.NoError(t, err)
	diff, err := qualitygate.NewMinPercentage(90)
	require.NoError(t, err)

	assert.True(t, lhs.Equal(rhs))
	assert.False(t, lhs.Equal(diff))
	assert.False(t, lhs.Equal(qualitygate.NewZeroTolerance()))
}

func TestForbiddenListEqualOrderIndependent(t *testing.T) {
	lhs, err := qualitygate.NewForbiddenList([]string{"GPL", "AGPL"})
	require.NoError(t, err)
	rhs, err := qualitygate.NewForbiddenList([]string{"AGPL", "GPL"})
	require.NoError(t, err)

	assert.True(t, lhs.Equal(rhs))
	assert.True(t, rhs.Equal(lhs))
}

func TestForbiddenListEqualDifferentSet(t *testing.T) {
	lhs, err := qualitygate.NewForbiddenList([]string{"GPL", "AGPL"})
	require.NoError(t, err)
	rhs, err := qualitygate.NewForbiddenList([]string{"GPL"})
	require.NoError(t, err)
	alt, err := qualitygate.NewForbiddenList([]string{"GPL", "MIT"})
	require.NoError(t, err)

	assert.False(t, lhs.Equal(rhs))
	assert.False(t, lhs.Equal(alt))
	assert.False(t, lhs.Equal(qualitygate.NewZeroTolerance()))
}

func TestMaxViolationsEqual(t *testing.T) {
	lhs, err := qualitygate.NewMaxViolations(5)
	require.NoError(t, err)
	rhs, err := qualitygate.NewMaxViolations(5)
	require.NoError(t, err)
	alt, err := qualitygate.NewMaxViolations(10)
	require.NoError(t, err)

	assert.True(t, lhs.Equal(rhs))
	assert.False(t, lhs.Equal(alt))
	assert.False(t, lhs.Equal(qualitygate.NewZeroTolerance()))
}

func TestMinNewCodeCoverageEqual(t *testing.T) {
	lhs, err := qualitygate.NewMinNewCodeCoverage(80)
	require.NoError(t, err)
	rhs, err := qualitygate.NewMinNewCodeCoverage(80)
	require.NoError(t, err)
	diff, err := qualitygate.NewMinNewCodeCoverage(70)
	require.NoError(t, err)

	assert.True(t, lhs.Equal(rhs))
	assert.False(t, lhs.Equal(diff))
	assert.False(t, lhs.Equal(qualitygate.NewZeroTolerance()))
}
