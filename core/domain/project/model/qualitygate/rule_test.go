package qualitygate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

func TestNewQualityGateRuleValid(t *testing.T) {
	countStrategy, err := qualitygate.NewCountBySeverity(5, 10, 20)
	require.NoError(t, err)
	zeroStrategy := qualitygate.NewZeroTolerance()
	percentStrategy, err := qualitygate.NewMinPercentage(80)
	require.NoError(t, err)
	forbiddenStrategy, err := qualitygate.NewForbiddenList([]string{"GPL-3.0"})
	require.NoError(t, err)
	maxViolationsStrategy, err := qualitygate.NewMaxViolations(0)
	require.NoError(t, err)

	tests := []struct {
		name     string
		subtype  evidence.Subtype
		strategy qualitygate.Strategy
	}{
		{
			name:     "countBySeverityWithCodeQuality",
			subtype:  evidence.SubtypeCodeQuality,
			strategy: countStrategy,
		},
		{
			name:     "zeroToleranceWithSAST",
			subtype:  evidence.SubtypeSAST,
			strategy: zeroStrategy,
		},
		{
			name:     "countBySeverityWithComplexity",
			subtype:  evidence.SubtypeComplexity,
			strategy: countStrategy,
		},
		{
			name:     "countBySeverityWithSecrets",
			subtype:  evidence.SubtypeSecrets,
			strategy: countStrategy,
		},
		{
			name:     "maxViolationsWithArchitecture",
			subtype:  evidence.SubtypeArchitecture,
			strategy: maxViolationsStrategy,
		},
		{
			name:     "minPercentageWithCoverage",
			subtype:  evidence.SubtypeCoverage,
			strategy: percentStrategy,
		},
		{
			name:     "forbiddenListWithLicense",
			subtype:  evidence.SubtypeLicense,
			strategy: forbiddenStrategy,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			rule, err := qualitygate.NewRule(tcase.subtype, tcase.strategy)
			require.NoError(t, err)
			assert.Equal(t, tcase.subtype, rule.Subtype())
			assert.NotNil(t, rule.Strategy())
		})
	}
}

func TestNewQualityGateRuleIncompatible(t *testing.T) {
	countStrategy, err := qualitygate.NewCountBySeverity(5, 10, 20)
	require.NoError(t, err)
	zeroStrategy := qualitygate.NewZeroTolerance()
	percentStrategy, err := qualitygate.NewMinPercentage(80)
	require.NoError(t, err)
	forbiddenStrategy, err := qualitygate.NewForbiddenList([]string{"GPL-3.0"})
	require.NoError(t, err)
	maxViolationsStrategy, err := qualitygate.NewMaxViolations(0)
	require.NoError(t, err)

	tests := []struct {
		name     string
		subtype  evidence.Subtype
		strategy qualitygate.Strategy
	}{
		{
			name:     "countBySeverityWithCoverage",
			subtype:  evidence.SubtypeCoverage,
			strategy: countStrategy,
		},
		{
			name:     "countBySeverityWithLicense",
			subtype:  evidence.SubtypeLicense,
			strategy: countStrategy,
		},
		{
			name:     "countBySeverityWithArchitecture",
			subtype:  evidence.SubtypeArchitecture,
			strategy: countStrategy,
		},
		{
			name:     "zeroToleranceWithCoverage",
			subtype:  evidence.SubtypeCoverage,
			strategy: zeroStrategy,
		},
		{
			name:     "zeroToleranceWithLicense",
			subtype:  evidence.SubtypeLicense,
			strategy: zeroStrategy,
		},
		{
			name:     "zeroToleranceWithArchitecture",
			subtype:  evidence.SubtypeArchitecture,
			strategy: zeroStrategy,
		},
		{
			name:     "minPercentageWithCodeQuality",
			subtype:  evidence.SubtypeCodeQuality,
			strategy: percentStrategy,
		},
		{
			name:     "minPercentageWithSAST",
			subtype:  evidence.SubtypeSAST,
			strategy: percentStrategy,
		},
		{
			name:     "forbiddenListWithCodeQuality",
			subtype:  evidence.SubtypeCodeQuality,
			strategy: forbiddenStrategy,
		},
		{
			name:     "forbiddenListWithCoverage",
			subtype:  evidence.SubtypeCoverage,
			strategy: forbiddenStrategy,
		},
		{
			name:     "maxViolationsWithCodeQuality",
			subtype:  evidence.SubtypeCodeQuality,
			strategy: maxViolationsStrategy,
		},
		{
			name:     "maxViolationsWithCoverage",
			subtype:  evidence.SubtypeCoverage,
			strategy: maxViolationsStrategy,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			_, err := qualitygate.NewRule(tcase.subtype, tcase.strategy)
			require.Error(t, err)
			assert.ErrorIs(t, err, qualitygate.ErrInvalidRule)
		})
	}
}

func TestNewQualityGateRuleNilStrategy(t *testing.T) {
	_, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, qualitygate.ErrInvalidRule)
}

func TestNewRuleWithMinResolved(t *testing.T) {
	rule, err := qualitygate.NewRule(
		evidence.SubtypeCodeQuality,
		qualitygate.NewZeroTolerance(),
		qualitygate.WithMinResolved(5),
	)
	require.NoError(t, err)
	require.NotNil(t, rule.MinResolved())
	assert.Equal(t, 5, *rule.MinResolved())
	assert.Nil(t, rule.MinDelta())
}

func TestNewRuleWithMinDelta(t *testing.T) {
	pct, err := qualitygate.NewMinPercentage(80)
	require.NoError(t, err)
	rule, err := qualitygate.NewRule(
		evidence.SubtypeCoverage,
		pct,
		qualitygate.WithMinDelta(0),
	)
	require.NoError(t, err)
	require.NotNil(t, rule.MinDelta())
	assert.InDelta(t, 0.0, *rule.MinDelta(), 0.001)
	assert.Nil(t, rule.MinResolved())
}

func TestNewRuleWithoutDeltaFieldsReturnsNil(t *testing.T) {
	rule, err := qualitygate.NewRule(
		evidence.SubtypeCodeQuality,
		qualitygate.NewZeroTolerance(),
	)
	require.NoError(t, err)
	assert.Nil(t, rule.MinResolved())
	assert.Nil(t, rule.MinDelta())
}

type unknownStrategy struct{}

func (u unknownStrategy) Evaluate(evidence.Content) qualitygate.Outcome {
	return qualitygate.NewOutcome(true, "")
}

func (u unknownStrategy) Equal(qualitygate.Strategy) bool { return false }

func TestNewRuleRejectsUnknownStrategyType(t *testing.T) {
	_, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, unknownStrategy{})
	require.Error(t, err)
	assert.ErrorIs(t, err, qualitygate.ErrInvalidRule)
}

func TestNewRuleMinNewCodeCoverageIncompatible(t *testing.T) {
	ncc, err := qualitygate.NewMinNewCodeCoverage(80)
	require.NoError(t, err)

	_, err = qualitygate.NewRule(evidence.SubtypeCoverage, ncc)
	require.Error(t, err)
	assert.ErrorIs(t, err, qualitygate.ErrInvalidRule)
}

func TestRuleEqualDifferentSubtypes(t *testing.T) {
	ruleA, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, qualitygate.NewZeroTolerance())
	require.NoError(t, err)
	ruleB, err := qualitygate.NewRule(evidence.SubtypeSAST, qualitygate.NewZeroTolerance())
	require.NoError(t, err)

	assert.False(t, ruleA.Equal(ruleB))
}

func TestRuleEqualDifferentMinDelta(t *testing.T) {
	pct, err := qualitygate.NewMinPercentage(80)
	require.NoError(t, err)
	ruleA, err := qualitygate.NewRule(evidence.SubtypeCoverage, pct, qualitygate.WithMinDelta(0))
	require.NoError(t, err)
	ruleB, err := qualitygate.NewRule(evidence.SubtypeCoverage, pct, qualitygate.WithMinDelta(5))
	require.NoError(t, err)
	ruleC, err := qualitygate.NewRule(evidence.SubtypeCoverage, pct)
	require.NoError(t, err)

	assert.False(t, ruleA.Equal(ruleB), "different minDelta values")
	assert.False(t, ruleA.Equal(ruleC), "minDelta vs none")
}

func TestNewRuleRejectsNegativeMinResolved(t *testing.T) {
	_, err := qualitygate.NewRule(
		evidence.SubtypeCodeQuality,
		qualitygate.NewZeroTolerance(),
		qualitygate.WithMinResolved(-1),
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, qualitygate.ErrInvalidRule)
}
