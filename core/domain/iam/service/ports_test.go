package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

type fakeTenantRepo struct{}

func (fakeTenantRepo) Save(context.Context, tenant.Tenant) error { return nil }
func (fakeTenantRepo) ByID(context.Context, tenant.TenantID) (tenant.Tenant, error) {
	return tenant.Tenant{}, nil
}
func (fakeTenantRepo) BySlug(context.Context, tenant.Slug) (tenant.Tenant, error) {
	return tenant.Tenant{}, nil
}

type fakeUserRepo struct{}

func (fakeUserRepo) Save(context.Context, user.User) error                { return nil }
func (fakeUserRepo) ByID(context.Context, user.UserID) (user.User, error) { return user.User{}, nil }
func (fakeUserRepo) ByEmail(context.Context, tenant.TenantID, user.Email) (user.User, error) {
	return user.User{}, nil
}
func (fakeUserRepo) CountByTenant(context.Context, tenant.TenantID) (int, error) { return 0, nil }

type fakeSessionRepo struct{}

func (fakeSessionRepo) Save(context.Context, session.Session) error { return nil }
func (fakeSessionRepo) ByTokenHash(context.Context, session.TokenHash) (session.Session, error) {
	return session.Session{}, nil
}
func (fakeSessionRepo) DeleteAllForUser(context.Context, user.UserID) error     { return nil }
func (fakeSessionRepo) DeleteExpired(context.Context, time.Time) (int64, error) { return 0, nil }

type fakeAPITokenRepo struct{}

func (fakeAPITokenRepo) Save(context.Context, apitoken.APIToken) error { return nil }
func (fakeAPITokenRepo) ByID(context.Context, apitoken.APITokenID) (apitoken.APIToken, error) {
	return apitoken.APIToken{}, nil
}
func (fakeAPITokenRepo) ByTokenHash(context.Context, apitoken.SecretHash) (apitoken.APIToken, error) {
	return apitoken.APIToken{}, nil
}
func (fakeAPITokenRepo) ListByUser(context.Context, user.UserID) ([]apitoken.APIToken, error) {
	return nil, nil
}

type fakeHasher struct{}

func (fakeHasher) Hash(string) (user.PasswordHash, error)         { return user.PasswordHash{}, nil }
func (fakeHasher) Verify(string, user.PasswordHash) (bool, error) { return false, nil }

type fakeSecretGen struct{}

func (fakeSecretGen) NewSessionToken() (session.Token, error) { return session.Token{}, nil }
func (fakeSecretGen) NewAPITokenSecret() (apitoken.Secret, error) {
	return apitoken.Secret{}, nil
}

func TestPortsAreSatisfiedByImplementations(t *testing.T) {
	var _ service.TenantRepository = fakeTenantRepo{}
	var _ service.UserRepository = fakeUserRepo{}
	var _ service.SessionRepository = fakeSessionRepo{}
	var _ service.APITokenRepository = fakeAPITokenRepo{}
	var _ service.PasswordHasher = fakeHasher{}
	var _ service.SecretGenerator = fakeSecretGen{}
}
