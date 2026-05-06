package model_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/project/model"
)

func TestNewProjectIDWrapsUUID(t *testing.T) {
	u := uuid.New()
	id := model.NewProjectID(u)

	assert.Equal(t, u, id.UUID())
	assert.Equal(t, u.String(), id.String())
}

func TestProjectIDEqualByUnderlyingUUID(t *testing.T) {
	u := uuid.New()
	a := model.NewProjectID(u)
	b := model.NewProjectID(u)
	assert.True(t, a.Equal(b))

	c := model.NewProjectID(uuid.New())
	assert.False(t, a.Equal(c))
}

func TestParseProjectIDAcceptsCanonicalUUID(t *testing.T) {
	u := uuid.New()
	id, err := model.ParseProjectID(u.String())
	require.NoError(t, err)
	assert.Equal(t, u, id.UUID())
}

func TestParseProjectIDRejectsNonUUID(t *testing.T) {
	for _, raw := range []string{"", "   ", "not-a-uuid", "proj-1", "550e8400-e29b-41d4-a716", uuid.Nil.String()} {
		_, err := model.ParseProjectID(raw)
		require.Error(t, err)
		assert.ErrorIs(t, err, model.ErrInvalidProject)
	}
}
