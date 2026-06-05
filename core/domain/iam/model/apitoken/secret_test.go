package apitoken_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
)

const (
	validSecret       = "gav_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	validAPITokenHash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
)

func TestNewAPITokenSecret(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid gav prefix + 43-char body", value: validSecret},
		{name: "empty rejected", value: "", wantErr: true},
		{name: "missing gav prefix rejected", value: "abcdefghij_KLMNOPQRSTuvwxyz0123456789-ABCD_EF___"[:47], wantErr: true},
		{name: "wrong prefix rejected", value: "tok_" + validSecret[4:], wantErr: true},
		{name: "too short rejected", value: "gav_abc", wantErr: true},
		{name: "too long rejected", value: validSecret + "00", wantErr: true},
		{name: "non-base64 body rejected", value: "gav_" + strings.Repeat("!", 43), wantErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			sec, err := apitoken.NewSecret(tcase.value)
			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, apitoken.ErrInvalid)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tcase.value, sec.String())
		})
	}
}

func TestAPITokenSecretPrefix(t *testing.T) {
	s, err := apitoken.NewSecret(validSecret)
	require.NoError(t, err)
	prefix := s.Prefix()
	assert.Equal(t, "gav_AAAAAAAA", prefix, "Prefix exposes the leading 12 characters for UI identification")
}

func TestAPITokenSecretPrefixShortValue(t *testing.T) {
	short := apitoken.Secret{}
	assert.Equal(t, "", short.Prefix())
}

func TestAPITokenSecretEqual(t *testing.T) {
	a, _ := apitoken.NewSecret(validSecret)
	b, _ := apitoken.NewSecret(validSecret)
	assert.True(t, a.Equal(b))
}

func TestNewAPITokenHash(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid 64-char lowercase hex", value: validAPITokenHash},
		{name: "uppercase normalised", value: strings.ToUpper(validAPITokenHash)},
		{name: "empty rejected", value: "", wantErr: true},
		{name: "too short rejected", value: "abc", wantErr: true},
		{name: "non-hex rejected", value: strings.Repeat("g", 64), wantErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			hash, err := apitoken.NewSecretHash(tcase.value)
			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, apitoken.ErrInvalid)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, strings.ToLower(tcase.value), hash.String())
		})
	}
}

func TestHashAPITokenSecret(t *testing.T) {
	s, _ := apitoken.NewSecret(validSecret)
	h1 := apitoken.HashSecret(s)
	h2 := apitoken.HashSecret(s)
	assert.True(t, h1.Equal(h2))
	assert.Equal(t, 64, len(h1.String()))
}
