package model_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model"
)

func TestNewCaseFileIDWrapsUUID(t *testing.T) {
	u := uuid.New()
	id := model.NewCaseFileID(u)

	assert.Equal(t, u, id.UUID(), "UUID() must return the underlying value")
	assert.Equal(t, u.String(), id.String(), "String() must match the underlying UUID string")
}

func TestCaseFileIDEqualByUnderlyingUUID(t *testing.T) {
	u := uuid.New()
	a := model.NewCaseFileID(u)
	b := model.NewCaseFileID(u)
	assert.True(t, a.Equal(b), "two IDs wrapping the same UUID are equal")

	c := model.NewCaseFileID(uuid.New())
	assert.False(t, a.Equal(c), "different UUIDs are unequal")
}

func TestParseCaseFileIDAcceptsCanonicalUUID(t *testing.T) {
	u := uuid.New()
	id, err := model.ParseCaseFileID(u.String())
	require.NoError(t, err)
	assert.Equal(t, u, id.UUID())
}

func TestParseCaseFileIDRejectsNonUUID(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "whitespace", value: "   "},
		{name: "garbage", value: "not-a-uuid"},
		{name: "old-style semantic id", value: "case-1"},
		{name: "truncated uuid", value: "550e8400-e29b-41d4-a716"},
		{name: "nil uuid", value: uuid.Nil.String()},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			_, err := model.ParseCaseFileID(tcase.value)
			require.Error(t, err)
			assert.ErrorIs(t, err, model.ErrInvalidCaseFile)
		})
	}
}
