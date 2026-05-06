package qualitygate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

func TestNewQualityGateValid(t *testing.T) {
	rules := buildTwoRules(t)
	gate, err := qualitygate.NewGate(rules)
	require.NoError(t, err)
	assert.Len(t, gate.Rules(), 2)
}

func TestNewQualityGateEmptyRulesValid(t *testing.T) {
	gate, err := qualitygate.NewGate(nil)
	require.NoError(t, err)
	assert.Empty(t, gate.Rules())
}

func TestNewQualityGateDuplicateSubtype(t *testing.T) {
	strategy := qualitygate.NewZeroTolerance()
	rule1, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, strategy)
	require.NoError(t, err)

	pctStrategy, err := qualitygate.NewMinPercentage(80)
	require.NoError(t, err)
	rule2, err := qualitygate.NewRule(evidence.SubtypeCoverage, pctStrategy)
	require.NoError(t, err)

	_, err = qualitygate.NewGate([]qualitygate.Rule{rule1, rule1})
	require.Error(t, err)
	assert.ErrorIs(t, err, qualitygate.ErrInvalidGate)

	_, err = qualitygate.NewGate([]qualitygate.Rule{rule1, rule2})
	require.NoError(t, err)
}

func TestQualityGateRulesDefensiveCopy(t *testing.T) {
	rules := buildTwoRules(t)
	gate, err := qualitygate.NewGate(rules)
	require.NoError(t, err)

	returned := gate.Rules()
	returned[0] = qualitygate.Rule{}
	assert.NotEqual(t, returned[0], gate.Rules()[0])
}

func TestQualityGateRuleForSubtype(t *testing.T) {
	rules := buildTwoRules(t)
	gate, err := qualitygate.NewGate(rules)
	require.NoError(t, err)

	found, ok := gate.RuleForSubtype(evidence.SubtypeCodeQuality)
	assert.True(t, ok)
	assert.Equal(t, evidence.SubtypeCodeQuality, found.Subtype())

	_, ok = gate.RuleForSubtype(evidence.SubtypeSAST)
	assert.False(t, ok)
}

func buildTwoRules(t *testing.T) []qualitygate.Rule {
	t.Helper()
	strategy := qualitygate.NewZeroTolerance()
	rule1, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, strategy)
	require.NoError(t, err)

	pctStrategy, err := qualitygate.NewMinPercentage(80)
	require.NoError(t, err)
	rule2, err := qualitygate.NewRule(evidence.SubtypeCoverage, pctStrategy)
	require.NoError(t, err)

	return []qualitygate.Rule{rule1, rule2}
}

func TestQualityGateEqualSameRulesSameOrder(t *testing.T) {
	lhs, err := qualitygate.NewGate(buildTwoRules(t))
	require.NoError(t, err)
	rhs, err := qualitygate.NewGate(buildTwoRules(t))
	require.NoError(t, err)

	assert.True(t, lhs.Equal(rhs))
	assert.True(t, rhs.Equal(lhs))
}

func TestQualityGateEqualSameRulesDifferentOrder(t *testing.T) {
	rules := buildTwoRules(t)
	lhs, err := qualitygate.NewGate(rules)
	require.NoError(t, err)
	reversed := []qualitygate.Rule{rules[1], rules[0]}
	rhs, err := qualitygate.NewGate(reversed)
	require.NoError(t, err)

	assert.True(t, lhs.Equal(rhs))
}

func TestQualityGateEqualDifferentThreshold(t *testing.T) {
	lhs := buildCoverageGate(t, 80)
	rhs := buildCoverageGate(t, 90)

	assert.False(t, lhs.Equal(rhs))
	assert.False(t, rhs.Equal(lhs))
}

func TestQualityGateEqualExtraRule(t *testing.T) {
	lhs, err := qualitygate.NewGate(buildTwoRules(t))
	require.NoError(t, err)
	rules := buildTwoRules(t)
	rhs, err := qualitygate.NewGate([]qualitygate.Rule{rules[0]})
	require.NoError(t, err)

	assert.False(t, lhs.Equal(rhs))
	assert.False(t, rhs.Equal(lhs))
}

func TestQualityGateEqualEmptyVsEmpty(t *testing.T) {
	lhs, err := qualitygate.NewGate(nil)
	require.NoError(t, err)
	rhs, err := qualitygate.NewGate(nil)
	require.NoError(t, err)

	assert.True(t, lhs.Equal(rhs))
}

func TestQualityGateEqualDifferentDeltaFields(t *testing.T) {
	ruleA, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, qualitygate.NewZeroTolerance(), qualitygate.WithMinResolved(5))
	require.NoError(t, err)
	ruleB, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, qualitygate.NewZeroTolerance(), qualitygate.WithMinResolved(10))
	require.NoError(t, err)
	ruleC, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, qualitygate.NewZeroTolerance())
	require.NoError(t, err)

	lhs, err := qualitygate.NewGate([]qualitygate.Rule{ruleA})
	require.NoError(t, err)
	rhs, err := qualitygate.NewGate([]qualitygate.Rule{ruleB})
	require.NoError(t, err)
	c, err := qualitygate.NewGate([]qualitygate.Rule{ruleC})
	require.NoError(t, err)

	assert.False(t, lhs.Equal(rhs), "different minResolved values")
	assert.False(t, lhs.Equal(c), "minResolved vs none")
	assert.True(t, lhs.Equal(lhs), "same gate")
}

func TestQualityGateEqualSameLengthDifferentSubtypes(t *testing.T) {
	ruleA, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, qualitygate.NewZeroTolerance())
	require.NoError(t, err)
	ruleB, err := qualitygate.NewRule(evidence.SubtypeSAST, qualitygate.NewZeroTolerance())
	require.NoError(t, err)

	lhs, err := qualitygate.NewGate([]qualitygate.Rule{ruleA})
	require.NoError(t, err)
	rhs, err := qualitygate.NewGate([]qualitygate.Rule{ruleB})
	require.NoError(t, err)

	assert.False(t, lhs.Equal(rhs))
}

func buildCoverageGate(t *testing.T, min float64) qualitygate.Gate {
	t.Helper()
	strategy, err := qualitygate.NewMinPercentage(min)
	require.NoError(t, err)
	rule, err := qualitygate.NewRule(evidence.SubtypeCoverage, strategy)
	require.NoError(t, err)
	gate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)
	return gate
}
