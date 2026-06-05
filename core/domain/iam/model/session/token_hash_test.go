package session_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
)

const validHash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestNewSessionTokenHash(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "valid 64-char lowercase hex", value: validHash, want: validHash},
		{name: "uppercase normalised to lowercase", value: strings.ToUpper(validHash), want: validHash},
		{name: "empty rejected", value: "", wantErr: true},
		{name: "too short rejected", value: "abc", wantErr: true},
		{name: "too long rejected", value: validHash + "00", wantErr: true},
		{name: "non-hex rejected", value: strings.Repeat("g", 64), wantErr: true},
		{name: "whitespace rejected", value: "  " + validHash + "  ", wantErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			hash, err := session.NewTokenHash(tcase.value)
			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, session.ErrInvalid)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tcase.want, hash.String())
		})
	}
}

func TestHashSessionToken(t *testing.T) {
	tok, err := session.NewToken(validSessionToken[:43])
	require.NoError(t, err)

	h1 := session.HashToken(tok)
	h2 := session.HashToken(tok)
	assert.True(t, h1.Equal(h2), "hashing the same token twice must yield the same hash")
	assert.Equal(t, 64, len(h1.String()), "SHA-256 hex digest is 64 characters")
}

func TestSessionTokenHashEqual(t *testing.T) {
	a, _ := session.NewTokenHash(validHash)
	b, _ := session.NewTokenHash(strings.ToUpper(validHash))
	assert.True(t, a.Equal(b))
}
