package createuser

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

func TestNewHandlerPanicsOnNilTenants(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(nil, &stubUserRepo{}, &stubHasher{})
	})
}

func TestNewHandlerPanicsOnNilUsers(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(&stubTenantRepo{}, nil, &stubHasher{})
	})
}

func TestNewHandlerPanicsOnNilHasher(t *testing.T) {
	assert.Panics(t, func() {
		NewHandler(&stubTenantRepo{}, &stubUserRepo{}, nil)
	})
}

func TestExecuteReturnsErrorOnInvalidTenantID(t *testing.T) {
	tn := seedInternalTenant(t)
	handler := NewHandler(&stubTenantRepo{tenant: tn}, &stubUserRepo{}, &stubHasher{})
	cmd := Command{tenantID: "not-a-uuid", email: "a@b.com", displayName: "Alice", role: "admin", plainPassword: "hunter22", occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, tenant.ErrInvalidTenant)
}

func TestExecuteReturnsErrorOnInvalidEmail(t *testing.T) {
	tn := seedInternalTenant(t)
	handler := NewHandler(&stubTenantRepo{tenant: tn}, &stubUserRepo{}, &stubHasher{})
	cmd := Command{tenantID: tn.ID().String(), email: "not-an-email", displayName: "Alice", role: "admin", plainPassword: "hunter22", occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrInvalidEmail)
}

func TestExecuteReturnsErrorOnInvalidRole(t *testing.T) {
	tn := seedInternalTenant(t)
	handler := NewHandler(&stubTenantRepo{tenant: tn}, &stubUserRepo{}, &stubHasher{})
	cmd := Command{tenantID: tn.ID().String(), email: "a@b.com", displayName: "Alice", role: "emperor", plainPassword: "hunter22", occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrInvalidUser)
}

func TestExecuteReturnsErrorOnHashFailure(t *testing.T) {
	tn := seedInternalTenant(t)
	hasher := &stubHasher{hashErr: errors.New("hash broken")}
	handler := NewHandler(&stubTenantRepo{tenant: tn}, &stubUserRepo{}, hasher)
	cmd := Command{tenantID: tn.ID().String(), email: "a@b.com", displayName: "Alice", role: "admin", plainPassword: "hunter22", occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hash password")
}

func TestExecuteReturnsErrorOnNewUserDomainFailure(t *testing.T) {
	tn := seedInternalTenant(t)
	hash, _ := user.NewPasswordHash("$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g")
	hasher := &stubHasher{hashResult: hash}
	handler := NewHandler(&stubTenantRepo{tenant: tn}, &stubUserRepo{}, hasher)
	cmd := Command{tenantID: tn.ID().String(), email: "a@b.com", displayName: "", role: "admin", plainPassword: "hunter22", occurredAt: internalTestTime}
	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "new user")
}

func seedInternalTenant(t *testing.T) tenant.Tenant {
	t.Helper()
	slug, err := tenant.NewSlug("acme")
	require.NoError(t, err)
	tn, err := tenant.NewTenant(slug, "Acme", internalTestTime)
	require.NoError(t, err)
	tn.ClearEvents()
	return tn
}

type stubTenantRepo struct {
	tenant tenant.Tenant
}

func (r *stubTenantRepo) Save(_ context.Context, _ tenant.Tenant) error { return nil }
func (r *stubTenantRepo) ByID(_ context.Context, _ tenant.TenantID) (tenant.Tenant, error) {
	return r.tenant, nil
}
func (r *stubTenantRepo) BySlug(_ context.Context, _ tenant.Slug) (tenant.Tenant, error) {
	return r.tenant, nil
}

type stubUserRepo struct {
	saveErr error
}

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
	hashResult user.PasswordHash
	hashErr    error
}

func (h *stubHasher) Hash(_ string) (user.PasswordHash, error)            { return h.hashResult, h.hashErr }
func (h *stubHasher) Verify(_ string, _ user.PasswordHash) (bool, error) { return false, nil }
