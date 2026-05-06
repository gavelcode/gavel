package archpolicy_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/project/model/archpolicy"
)

const mutatedValue = "mutated"

func TestNewLayer(t *testing.T) {
	tests := []struct {
		name     string
		lName    string
		patterns []string
		wantErr  bool
	}{
		{
			name:     "shouldCreateValidLayer",
			lName:    "domain",
			patterns: []string{"internal/domain/..."},
		},
		{
			name:     "shouldCreateLayerWithMultiplePatterns",
			lName:    "domain",
			patterns: []string{"internal/domain/...", "pkg/domain/..."},
		},
		{
			name:     "shouldRejectEmptyName",
			lName:    "",
			patterns: []string{"internal/domain/..."},
			wantErr:  true,
		},
		{
			name:     "shouldRejectBlankName",
			lName:    "   ",
			patterns: []string{"internal/domain/..."},
			wantErr:  true,
		},
		{
			name:     "shouldRejectNoPatterns",
			lName:    "domain",
			patterns: nil,
			wantErr:  true,
		},
		{
			name:     "shouldRejectEmptyPattern",
			lName:    "domain",
			patterns: []string{"internal/domain/...", ""},
			wantErr:  true,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			layer, err := archpolicy.NewLayer(tcase.lName, tcase.patterns)
			if tcase.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tcase.lName, layer.Name())
			assert.Equal(t, tcase.patterns, layer.Patterns())
		})
	}
}

func TestLayerPatternsDefensiveCopy(t *testing.T) {
	original := []string{"internal/domain/..."}
	layer, err := archpolicy.NewLayer("domain", original)
	require.NoError(t, err)

	original[0] = mutatedValue
	assert.Equal(t, "internal/domain/...", layer.Patterns()[0])

	retrieved := layer.Patterns()
	retrieved[0] = mutatedValue
	assert.Equal(t, "internal/domain/...", layer.Patterns()[0])
}
