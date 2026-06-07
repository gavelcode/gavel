package issuetoken_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/iam/issuetoken"
	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	memiam "github.com/usegavel/gavel/core/infrastructure/iam/memory"
)

var (
	testTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)
)

func seedUser(t *testing.T, users *memiam.UserRepository, active bool) user.User {
	t.Helper()
	tenantID := tenant.NewTenantID(uuid.New())
	email, _ := user.NewEmail("alice@example.com")
	hash, _ := user.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g")
	foundUser, err := user.NewUser(tenantID, email, "Alice", user.RoleAdmin, hash, false, testTime)
	require.NoError(t, err)
	foundUser.ClearEvents()
	if !active {
		require.NoError(t, foundUser.Deactivate(testTime.Add(time.Minute)))
		foundUser.ClearEvents()
	}
	require.NoError(t, users.Save(context.Background(), foundUser))
	return foundUser
}

func TestExecuteIssuesToken(t *testing.T) {
	users := memiam.NewUserRepository()
	tokens := memiam.NewAPITokenRepository()
	secrets := memiam.NewFakeSecretGenerator()
	foundUser := seedUser(t, users, true)

	cmd, err := issuetoken.NewCommand(foundUser.ID().String(), "ci-bot", []string{"ingest", "read"}, testTime, testTime.Add(30*24*time.Hour))
	require.NoError(t, err)

	result, err := issuetoken.NewHandler(users, tokens, secrets).Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.NotEmpty(t, result.TokenID)
	assert.Equal(t, foundUser.ID().String(), result.UserID)
	assert.Equal(t, "ci-bot", result.Name)
	assert.Equal(t, []string{"ingest", "read"}, result.Scopes)
	assert.NotEmpty(t, result.PlainSecret, "plaintext must be returned exactly once")
	assert.NotEmpty(t, result.TokenPrefix)
	require.Len(t, result.Events, 1)
	assert.Equal(t, apitoken.EventNameIssued, result.Events[0].Name)

	id, _ := apitoken.ParseAPITokenID(result.TokenID)
	saved, err := tokens.ByID(context.Background(), id)
	require.NoError(t, err)
	assert.False(t, saved.IsRevoked())
}

func TestExecuteRejectsInactiveUser(t *testing.T) {
	users := memiam.NewUserRepository()
	tokens := memiam.NewAPITokenRepository()
	secrets := memiam.NewFakeSecretGenerator()
	foundUser := seedUser(t, users, false)

	cmd, _ := issuetoken.NewCommand(foundUser.ID().String(), "ci-bot", []string{"ingest"}, testTime, testTime.Add(time.Hour))
	_, err := issuetoken.NewHandler(users, tokens, secrets).Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, apitoken.ErrInvalid)
}

func TestExecuteRejectsMissingUser(t *testing.T) {
	users := memiam.NewUserRepository()
	tokens := memiam.NewAPITokenRepository()
	secrets := memiam.NewFakeSecretGenerator()

	cmd, _ := issuetoken.NewCommand(uuid.NewString(), "ci-bot", []string{"ingest"}, testTime, testTime.Add(time.Hour))
	_, err := issuetoken.NewHandler(users, tokens, secrets).Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestExecuteRejectsUnknownScope(t *testing.T) {
	users := memiam.NewUserRepository()
	tokens := memiam.NewAPITokenRepository()
	secrets := memiam.NewFakeSecretGenerator()
	foundUser := seedUser(t, users, true)

	cmd, _ := issuetoken.NewCommand(foundUser.ID().String(), "ci-bot", []string{"write"}, testTime, testTime.Add(time.Hour))
	_, err := issuetoken.NewHandler(users, tokens, secrets).Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, apitoken.ErrInvalid)
}

func TestNewCommandRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name string
		fn   func() (issuetoken.Command, error)
	}{
		{name: "empty user", fn: func() (issuetoken.Command, error) {
			return issuetoken.NewCommand("", "n", []string{"read"}, testTime, time.Time{})
		}},
		{name: "empty name", fn: func() (issuetoken.Command, error) {
			return issuetoken.NewCommand("u-1", "", []string{"read"}, testTime, time.Time{})
		}},
		{name: "empty scopes", fn: func() (issuetoken.Command, error) {
			return issuetoken.NewCommand("u-1", "n", nil, testTime, time.Time{})
		}},
		{name: "zero time", fn: func() (issuetoken.Command, error) {
			return issuetoken.NewCommand("u-1", "n", []string{"read"}, time.Time{}, time.Time{})
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.fn()
			require.Error(t, err)
			assert.ErrorIs(t, err, issuetoken.ErrInvalidCommand)
		})
	}
}
