package evidence_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
)

func TestNewEvidenceIDWrapsUUID(t *testing.T) {
	u := uuid.New()
	id := evidence.NewEvidenceID(u)

	assert.Equal(t, u, id.UUID())
	assert.Equal(t, u.String(), id.String())
}

func TestEvidenceIDEqualByUnderlyingUUID(t *testing.T) {
	u := uuid.New()
	a := evidence.NewEvidenceID(u)
	b := evidence.NewEvidenceID(u)
	assert.True(t, a.Equal(b))
	assert.False(t, a.Equal(evidence.NewEvidenceID(uuid.New())))
}

func TestParseEvidenceIDAcceptsCanonicalUUID(t *testing.T) {
	u := uuid.New()
	id, err := evidence.ParseEvidenceID(u.String())
	require.NoError(t, err)
	assert.Equal(t, u, id.UUID())
}

func TestParseEvidenceIDRejectsNonUUID(t *testing.T) {
	for _, raw := range []string{"", "   ", "evid-1", "garbage", uuid.Nil.String()} {
		_, err := evidence.ParseEvidenceID(raw)
		require.Error(t, err)
		assert.ErrorIs(t, err, evidence.ErrInvalidEvidence)
	}
}
