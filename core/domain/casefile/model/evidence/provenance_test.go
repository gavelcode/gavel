package evidence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
)

func TestProvenanceZeroValueIsCatalog(t *testing.T) {
	var zero evidence.Provenance

	assert.Equal(t, evidence.ProvenanceCatalog, zero)
	assert.Equal(t, "catalog", zero.String())
	assert.True(t, zero.GuaranteedByGavel())
}

func TestProvenanceBridgeIsNotGuaranteed(t *testing.T) {
	assert.Equal(t, "bridge", evidence.ProvenanceBridge.String())
	assert.False(t, evidence.ProvenanceBridge.GuaranteedByGavel())
	assert.True(t, evidence.ProvenanceBridge.IsBridge())
	assert.False(t, evidence.ProvenanceCatalog.IsBridge())
}

func TestNewProvenance(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    evidence.Provenance
		wantErr bool
	}{
		{name: "shouldParseCatalog", input: "catalog", want: evidence.ProvenanceCatalog},
		{name: "shouldParseBridge", input: "bridge", want: evidence.ProvenanceBridge},
		{name: "shouldDefaultEmptyToCatalog", input: "", want: evidence.ProvenanceCatalog},
		{name: "shouldRejectUnknown", input: "magic", wantErr: true},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			got, err := evidence.NewProvenance(tcase.input)

			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, evidence.ErrInvalidProvenance)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tcase.want, got)
		})
	}
}
