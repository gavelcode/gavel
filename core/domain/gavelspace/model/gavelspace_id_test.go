package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/gavelspace/model"
)

func TestNewGavelspaceName(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid name", value: "monorepo"},
		{name: "with hyphens and dots", value: "acme.payments-team"},
		{name: "empty rejected", value: "", wantErr: true},
		{name: "whitespace rejected", value: "   ", wantErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			gid, err := model.NewGavelspaceID(tcase.value)
			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, model.ErrInvalidGavelspace)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tcase.value, gid.String())
		})
	}
}

func TestGavelspaceNameEqualAndZero(t *testing.T) {
	a, _ := model.NewGavelspaceID("ns")
	b, _ := model.NewGavelspaceID("ns")
	c, _ := model.NewGavelspaceID("other")
	assert.True(t, a.Equal(b))
	assert.False(t, a.Equal(c))
}
