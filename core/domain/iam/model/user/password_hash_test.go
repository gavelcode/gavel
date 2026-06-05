package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

const validArgon2Hash = "$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g"

func TestNewPasswordHash(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid argon2id", value: validArgon2Hash},
		{name: "empty rejected", value: "", wantErr: true},
		{name: "whitespace rejected", value: "   ", wantErr: true},
		{name: "plaintext rejected", value: "hunter2", wantErr: true},
		{name: "bcrypt rejected", value: "$2a$12$abcdefghijklmnopqrstuvwxyz0123456789ABCDEF", wantErr: true},
		{name: "wrong algorithm rejected", value: "$argon2d$v=19$m=65536,t=3,p=4$c2FsdA$aGFzaA", wantErr: true},
		{name: "missing parts rejected", value: "$argon2id$v=19$m=65536,t=3,p=4$c2FsdA", wantErr: true},
		{name: "leading text rejected", value: "garbage$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$aGFzaA", wantErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			hash, err := user.NewPasswordHash(tcase.value)
			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, user.ErrInvalidUser)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tcase.value, hash.String())
		})
	}
}

func TestPasswordHashEqual(t *testing.T) {
	a, _ := user.NewPasswordHash(validArgon2Hash)
	b, _ := user.NewPasswordHash(validArgon2Hash)
	assert.True(t, a.Equal(b))
}
