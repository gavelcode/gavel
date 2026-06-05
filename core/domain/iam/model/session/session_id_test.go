package session_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/session"
)

func TestParseSessionID(t *testing.T) {
	valid := uuid.New()
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid", input: valid.String()},
		{name: "valid with whitespace", input: "  " + valid.String() + "  "},
		{name: "empty rejected", input: "", wantErr: true},
		{name: "whitespace rejected", input: "   ", wantErr: true},
		{name: "not a uuid rejected", input: "not-a-uuid", wantErr: true},
		{name: "nil uuid rejected", input: uuid.Nil.String(), wantErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			sid, err := session.ParseSessionID(tcase.input)
			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, session.ErrInvalid)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, valid, sid.UUID())
			assert.Equal(t, valid.String(), sid.String())
		})
	}
}

func TestSessionIDEqual(t *testing.T) {
	raw := uuid.New()
	a := session.NewSessionID(raw)
	b := session.NewSessionID(raw)
	c := session.NewSessionID(uuid.New())

	assert.True(t, a.Equal(b))
	assert.False(t, a.Equal(c))
}
