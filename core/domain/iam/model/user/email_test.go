package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "plain valid", value: "alice@example.com", want: "alice@example.com"},
		{name: "uppercase normalised", value: "Alice@Example.COM", want: "alice@example.com"},
		{name: "surrounding whitespace trimmed", value: "  bob@example.com  ", want: "bob@example.com"},
		{name: "subaddress kept", value: "alice+test@example.com", want: "alice+test@example.com"},
		{name: "dotted local part kept", value: "first.last@example.com", want: "first.last@example.com"},
		{name: "empty rejected", value: "", wantErr: true},
		{name: "whitespace-only rejected", value: "   ", wantErr: true},
		{name: "no at sign rejected", value: "alice.example.com", wantErr: true},
		{name: "missing local part rejected", value: "@example.com", wantErr: true},
		{name: "missing domain rejected", value: "alice@", wantErr: true},
		{name: "two at signs rejected", value: "a@b@c", wantErr: true},
		{name: "domain without dot rejected", value: "alice@localhost", wantErr: true},
		{name: "internal whitespace rejected", value: "alice @example.com", wantErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			email, err := user.NewEmail(tcase.value)
			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, user.ErrInvalidEmail)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tcase.want, email.String())
		})
	}
}

func TestEmailEqualAndZero(t *testing.T) {
	a, _ := user.NewEmail("alice@example.com")
	b, _ := user.NewEmail("ALICE@example.com")
	c, _ := user.NewEmail("bob@example.com")

	assert.True(t, a.Equal(b), "case-different inputs must normalise to equal Emails")
	assert.False(t, a.Equal(c))
}
