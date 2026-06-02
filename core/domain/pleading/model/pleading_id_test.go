package model_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/pleading/model"
)

func TestNewPleadingIDWrapsUUID(t *testing.T) {
	u := uuid.New()
	id := model.NewPleadingID(u)

	assert.Equal(t, u, id.UUID())
	assert.Equal(t, u.String(), id.String())
}

func TestPleadingIDEqualByUnderlyingUUID(t *testing.T) {
	u := uuid.New()
	a := model.NewPleadingID(u)
	b := model.NewPleadingID(u)
	assert.True(t, a.Equal(b))
	assert.False(t, a.Equal(model.NewPleadingID(uuid.New())))
}

func TestParsePleadingIDAcceptsCanonicalUUID(t *testing.T) {
	u := uuid.New()
	id, err := model.ParsePleadingID(u.String())
	require.NoError(t, err)
	assert.Equal(t, u, id.UUID())
}

func TestParsePleadingIDRejectsNonUUID(t *testing.T) {
	for _, raw := range []string{"", "   ", "p-1", "garbage", uuid.Nil.String()} {
		_, err := model.ParsePleadingID(raw)
		require.Error(t, err)
		assert.ErrorIs(t, err, model.ErrInvalidPleading)
	}
}
