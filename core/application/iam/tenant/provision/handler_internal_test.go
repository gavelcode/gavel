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
	tenants := &stubTenantRepo{}
	users := &stubUserRepo{}
	hasher := &stubHasher{}

	assert.Panics(t, func() { NewHandler(nil, users, hasher) })
	assert.Panics(t, func() { NewHandler(tenants, nil, hasher) })
	assert.Panics(t, func() { NewHandler(tenants, users, nil) })
}

func TestExecuteReturnsErrorOnNewTenantDomainFailure(t *testing.T) {
	handler := NewHandler(&stubTenantRepo{}, &stubUserRepo{}, &stubHasher{hash: mustHash(t)})
	cmd := validCommand()
	cmd.displayName = ""

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "new tenant")
}

func TestExecuteReturnsErrorOnHashFailure(t *testing.T) {
	handler := NewHandler(&stubTenantRepo{}, &stubUserRepo{}, &stubHasher{err: errors.New("hasher broken")})

	_, err := handler.Execute(context.Background(), validCommand())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hash admin password")
}

func TestExecuteReturnsErrorOnNewUserDomainFailure(t *testing.T) {
	handler := NewHandler(&stubTenantRepo{}, &stubUserRepo{}, &stubHasher{hash: mustHash(t)})
	cmd := validCommand()
	cmd.adminDisplayName = ""

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "new admin user")
}

func TestExecuteReturnsErrorOnTenantSaveFailure(t *testing.T) {
	handler := NewHandler(&stubTenantRepo{saveErr: errors.New("tenant save broken")}, &stubUserRepo{}, &stubHasher{hash: mustHash(t)})

	_, err := handler.Execute(context.Background(), validCommand())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save tenant")
}

func TestExecuteReturnsErrorOnUserSaveFailure(t *testing.T) {
	handler := NewHandler(&stubTenantRepo{}, &stubUserRepo{saveErr: errors.New("user save broken")}, &stubHasher{hash: mustHash(t)})

	_, err := handler.Execute(context.Background(), validCommand())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save admin user")
}

type stubTenantRepo struct{ saveErr error }

func (r *stubTenantRepo) Save(_ context.Context, _ tenant.Tenant) error { return r.saveErr }
func (r *stubTenantRepo) ByID(_ context.Context, _ tenant.TenantID) (tenant.Tenant, error) {
	return tenant.Tenant{}, nil
}
func (r *stubTenantRepo) BySlug(_ context.Context, _ tenant.Slug) (tenant.Tenant, error) {
	return tenant.Tenant{}, nil
}

type stubUserRepo struct{ saveErr error }

func (r *stubUserRepo) Save(_ context.Context, _ user.User) error { return r.saveErr }
func (r *stubUserRepo) ByID(_ context.Context, _ user.UserID) (user.User, error) {
	return user.User{}, nil
}
func (r *stubUserRepo) ByEmail(_ context.Context, _ tenant.TenantID, _ user.Email) (user.User, error) {
	return user.User{}, nil
}
func (r *stubUserRepo) CountByTenant(_ context.Context, _ tenant.TenantID) (int, error) {
	return 0, nil
}

type stubHasher struct {
	hash user.PasswordHash
	err  error
}

func (h *stubHasher) Hash(_ string) (user.PasswordHash, error) { return h.hash, h.err }
func (h *stubHasher) Verify(_ string, _ user.PasswordHash) (bool, error) {
	return false, nil
}
