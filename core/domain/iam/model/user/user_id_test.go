package user_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

func TestNewUserIDWrapsUUID(t *testing.T) {
	u := uuid.New()
	id := user.NewUserID(u)
	assert.Equal(t, u, id.UUID())
	assert.Equal(t, u.String(), id.String())
}

func TestUserIDEqualByUnderlyingUUID(t *testing.T) {
	u := uuid.New()
	a := user.NewUserID(u)
	b := user.NewUserID(u)
	assert.True(t, a.Equal(b))
	assert.False(t, a.Equal(user.NewUserID(uuid.New())))
}

func TestParseUserIDAcceptsCanonicalUUID(t *testing.T) {
	u := uuid.New()
	id, err := user.ParseUserID(u.String())
	require.NoError(t, err)
	assert.Equal(t, u, id.UUID())
}

func TestParseUserIDRejectsNonUUID(t *testing.T) {
	for _, raw := range []string{"", "   ", "u-1", "garbage", uuid.Nil.String()} {
		_, err := user.ParseUserID(raw)
		require.Error(t, err)
		assert.ErrorIs(t, err, user.ErrInvalidUser)
	}
}
