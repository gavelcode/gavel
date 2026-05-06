package archpolicy_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/project/model/archpolicy"
)

func validLayers(t *testing.T) []archpolicy.Layer {
	t.Helper()
	domain, err := archpolicy.NewLayer("domain", []string{"internal/domain/..."})
	require.NoError(t, err)
	app, err := archpolicy.NewLayer("application", []string{"internal/application/..."})
	require.NoError(t, err)
	infra, err := archpolicy.NewLayer("infrastructure", []string{"internal/infrastructure/..."})
	require.NoError(t, err)
	return []archpolicy.Layer{domain, app, infra}
}

func validRule(t *testing.T) archpolicy.DenyRule {
	t.Helper()
	rule, err := archpolicy.NewDenyRule("domain-imports-nothing", "domain", []string{"application", "infrastructure"})
	require.NoError(t, err)
	return rule
}

func TestNewArchitecturePolicy(t *testing.T) {
	t.Run("shouldCreateValidPolicy", func(t *testing.T) {
		layers := validLayers(t)
		rule := validRule(t)

		policy, err := archpolicy.NewPolicy(layers, []archpolicy.DenyRule{rule}, true)
		require.NoError(t, err)
		assert.Len(t, policy.Layers(), 3)
		assert.Len(t, policy.DenyRules(), 1)
		assert.True(t, policy.DetectCycles())
	})

	t.Run("shouldCreatePolicyWithoutRules", func(t *testing.T) {
		layers := validLayers(t)

		policy, err := archpolicy.NewPolicy(layers, nil, false)
		require.NoError(t, err)
		assert.Len(t, policy.Layers(), 3)
		assert.Empty(t, policy.DenyRules())
		assert.False(t, policy.DetectCycles())
	})

	t.Run("shouldRejectNoLayers", func(t *testing.T) {
		_, err := archpolicy.NewPolicy(nil, nil, false)
		require.Error(t, err)
		assert.ErrorIs(t, err, archpolicy.ErrInvalidPolicy)
	})

	t.Run("shouldRejectDuplicateLayerNames", func(t *testing.T) {
		domain1, _ := archpolicy.NewLayer("domain", []string{"pkg1/..."})
		domain2, _ := archpolicy.NewLayer("domain", []string{"pkg2/..."})

		_, err := archpolicy.NewPolicy([]archpolicy.Layer{domain1, domain2}, nil, false)
		require.Error(t, err)
		assert.ErrorIs(t, err, archpolicy.ErrInvalidPolicy)
	})

	t.Run("shouldRejectRuleWithUnknownSourceLayer", func(t *testing.T) {
		layers := validLayers(t)
		badRule, _ := archpolicy.NewDenyRule("bad", "nonexistent", []string{"domain"})

		_, err := archpolicy.NewPolicy(layers, []archpolicy.DenyRule{badRule}, false)
		require.Error(t, err)
		assert.ErrorIs(t, err, archpolicy.ErrInvalidPolicy)
	})

	t.Run("shouldRejectRuleWithUnknownDenyLayer", func(t *testing.T) {
		layers := validLayers(t)
		badRule, _ := archpolicy.NewDenyRule("bad", "domain", []string{"nonexistent"})

		_, err := archpolicy.NewPolicy(layers, []archpolicy.DenyRule{badRule}, false)
		require.Error(t, err)
		assert.ErrorIs(t, err, archpolicy.ErrInvalidPolicy)
	})

	t.Run("shouldRejectRuleDenyingSelfSource", func(t *testing.T) {
		layers := validLayers(t)
		selfRule, _ := archpolicy.NewDenyRule("self-deny", "domain", []string{"domain"})

		_, err := archpolicy.NewPolicy(layers, []archpolicy.DenyRule{selfRule}, false)
		require.Error(t, err)
		assert.ErrorIs(t, err, archpolicy.ErrInvalidPolicy)
	})
}

func TestArchitecturePolicyDefensiveCopy(t *testing.T) {
	layers := validLayers(t)
	rule := validRule(t)

	policy, err := archpolicy.NewPolicy(layers, []archpolicy.DenyRule{rule}, true)
	require.NoError(t, err)

	retrievedLayers := policy.Layers()
	assert.Len(t, retrievedLayers, 3)
	retrievedLayers[0] = archpolicy.Layer{}
	assert.Equal(t, "domain", policy.Layers()[0].Name())

	retrievedRules := policy.DenyRules()
	assert.Len(t, retrievedRules, 1)
	retrievedRules[0] = archpolicy.DenyRule{}
	assert.Equal(t, "domain-imports-nothing", policy.DenyRules()[0].Name())
}
