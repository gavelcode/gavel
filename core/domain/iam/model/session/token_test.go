package session_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
)

const validSessionToken = "abcdefghij_KLMNOPQRSTuvwxyz0123456789-ABCD_EF"

func TestNewSessionToken(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid 43-char base64url", value: validSessionToken[:43]},
		{name: "empty rejected", value: "", wantErr: true},
		{name: "whitespace only rejected", value: "   ", wantErr: true},
		{name: "too short rejected", value: "abc", wantErr: true},
		{name: "too long rejected", value: strings.Repeat("a", 64), wantErr: true},
		{name: "contains slash rejected", value: "abcdefghij/KLMNOPQRSTuvwxyz0123456789-ABCD_EF"[:43], wantErr: true},
		{name: "contains plus rejected", value: "abcdefghij+KLMNOPQRSTuvwxyz0123456789-ABCD_EF"[:43], wantErr: true},
		{name: "contains equals padding rejected", value: "abcdefghij=KLMNOPQRSTuvwxyz0123456789-ABCD_EF"[:43], wantErr: true},
		{name: "contains whitespace rejected", value: "abcdefghij KLMNOPQRSTuvwxyz0123456789-ABCD_EF"[:43], wantErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			tok, err := session.NewToken(tcase.value)
			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, session.ErrInvalid)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tcase.value, tok.String())
		})
	}
}

func TestSessionTokenEqual(t *testing.T) {
	a, _ := session.NewToken(validSessionToken[:43])
	b, _ := session.NewToken(validSessionToken[:43])
	assert.True(t, a.Equal(b))
}
