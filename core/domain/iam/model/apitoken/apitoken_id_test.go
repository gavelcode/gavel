package apitoken_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
)

func TestNewAPITokenIDWrapsUUID(t *testing.T) {
	u := uuid.New()
	id := apitoken.NewAPITokenID(u)
	assert.Equal(t, u, id.UUID())
	assert.Equal(t, u.String(), id.String())
}

func TestAPITokenIDEqualByUnderlyingUUID(t *testing.T) {
	u := uuid.New()
	a := apitoken.NewAPITokenID(u)
	b := apitoken.NewAPITokenID(u)
	assert.True(t, a.Equal(b))
	assert.False(t, a.Equal(apitoken.NewAPITokenID(uuid.New())))
}

func TestParseAPITokenIDAcceptsCanonicalUUID(t *testing.T) {
	u := uuid.New()
	id, err := apitoken.ParseAPITokenID(u.String())
	require.NoError(t, err)
	assert.Equal(t, u, id.UUID())
}

func TestParseAPITokenIDRejectsNonUUID(t *testing.T) {
	for _, raw := range []string{"", "   ", "tok-1", "garbage", uuid.Nil.String()} {
		_, err := apitoken.ParseAPITokenID(raw)
		require.Error(t, err)
		assert.ErrorIs(t, err, apitoken.ErrInvalid)
	}
}
