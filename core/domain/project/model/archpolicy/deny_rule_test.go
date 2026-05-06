package archpolicy_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/project/model/archpolicy"
)

func TestNewDenyRule(t *testing.T) {
	tests := []struct {
		name    string
		rName   string
		source  string
		deny    []string
		wantErr bool
	}{
		{
			name:   "shouldCreateValidRule",
			rName:  "domain-imports-nothing",
			source: "domain",
			deny:   []string{"application", "infrastructure"},
		},
		{
			name:    "shouldRejectEmptyName",
			rName:   "",
			source:  "domain",
			deny:    []string{"application"},
			wantErr: true,
		},
		{
			name:    "shouldRejectEmptySource",
			rName:   "rule",
			source:  "",
			deny:    []string{"application"},
			wantErr: true,
		},
		{
			name:    "shouldRejectNoDenyTargets",
			rName:   "rule",
			source:  "domain",
			deny:    nil,
			wantErr: true,
		},
		{
			name:    "shouldRejectEmptyDenyTarget",
			rName:   "rule",
			source:  "domain",
			deny:    []string{"application", ""},
			wantErr: true,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			rule, err := archpolicy.NewDenyRule(tcase.rName, tcase.source, tcase.deny)
			if tcase.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tcase.rName, rule.Name())
			assert.Equal(t, tcase.source, rule.Source())
			assert.Equal(t, tcase.deny, rule.Deny())
		})
	}
}

func TestDenyRuleDenyDefensiveCopy(t *testing.T) {
	original := []string{"application", "infrastructure"}
	rule, err := archpolicy.NewDenyRule("rule", "domain", original)
	require.NoError(t, err)

	original[0] = mutatedValue
	assert.Equal(t, "application", rule.Deny()[0])

	retrieved := rule.Deny()
	retrieved[0] = mutatedValue
	assert.Equal(t, "application", rule.Deny()[0])
}
