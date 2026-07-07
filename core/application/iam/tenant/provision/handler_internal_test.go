package provision

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

var internalTestTime = time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC)

func validCommand() Command {
	return Command{
		slug:             "acme",
		displayName:      "Acme",
		adminEmail:       "admin@acme.com",
		adminDisplayName: "Administrator",
		adminPassword:    "s3cret-pass",
		occurredAt:       internalTestTime,
	}
}

func mustHash(t *testing.T) user.PasswordHash {
	t.Helper()
	hash, err := user.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$YWJjZGVmZ2hpamts$c29tZWtleXZhbHVlaGVy")
	require.NoError(t, err)
	return hash
}

func TestNewHandlerPanicsOnNilDependencies(t *testing.T) {
	provisioner := &stubProvisioner{}
	hasher := &stubHasher{}

	assert.Panics(t, func() { NewHandler(nil, hasher) })
	assert.Panics(t, func() { NewHandler(provisioner, nil) })
}

func TestExecuteReturnsErrorOnNewTenantDomainFailure(t *testing.T) {
	handler := NewHandler(&stubProvisioner{}, &stubHasher{hash: mustHash(t)})
	cmd := validCommand()
	cmd.displayName = ""

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "new tenant")
}

func TestExecuteReturnsErrorOnHashFailure(t *testing.T) {
	handler := NewHandler(&stubProvisioner{}, &stubHasher{err: errors.New("hasher broken")})

	_, err := handler.Execute(context.Background(), validCommand())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hash admin password")
}

func TestExecuteReturnsErrorOnNewUserDomainFailure(t *testing.T) {
	handler := NewHandler(&stubProvisioner{}, &stubHasher{hash: mustHash(t)})
	cmd := validCommand()
	cmd.adminDisplayName = ""

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "new admin user")
}

func TestExecuteReturnsErrorOnProvisionFailure(t *testing.T) {
	handler := NewHandler(&stubProvisioner{err: errors.New("provision broken")}, &stubHasher{hash: mustHash(t)})

	_, err := handler.Execute(context.Background(), validCommand())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "provision tenant")
}

type stubProvisioner struct{ err error }

func (p *stubProvisioner) Provision(_ context.Context, _ tenant.Tenant, _ user.User) error {
	return p.err
}

type stubHasher struct {
	hash user.PasswordHash
	err  error
}

func (h *stubHasher) Hash(_ string) (user.PasswordHash, error) { return h.hash, h.err }
func (h *stubHasher) Verify(_ string, _ user.PasswordHash) (bool, error) {
	return false, nil
}
