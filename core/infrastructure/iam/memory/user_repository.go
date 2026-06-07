package iam

import (
	"context"
	"fmt"
	"sync"

	tenantmodel "github.com/usegavel/gavel/core/domain/iam/model/tenant"
	usermodel "github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

var _ service.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	mu              sync.RWMutex
	byID            map[string]usermodel.User
	idByTenantEmail map[string]string
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		byID:            make(map[string]usermodel.User),
		idByTenantEmail: make(map[string]string),
	}
}

func tenantEmailKey(tenantID tenantmodel.TenantID, email usermodel.Email) string {
	return tenantID.String() + "|" + email.String()
}

func (r *UserRepository) Save(_ context.Context, user usermodel.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := tenantEmailKey(user.TenantID(), user.Email())
	if existingID, taken := r.idByTenantEmail[key]; taken && existingID != user.ID().String() {
		return fmt.Errorf("%w: %s", usermodel.ErrEmailAlreadyInUse, user.Email().String())
	}
	r.byID[user.ID().String()] = user
	r.idByTenantEmail[key] = user.ID().String()
	return nil
}

func (r *UserRepository) ByID(_ context.Context, id usermodel.UserID) (usermodel.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.byID[id.String()]
	if !ok {
		return usermodel.User{}, fmt.Errorf("%w: %s", usermodel.ErrUserNotFound, id.String())
	}
	return user, nil
}

func (r *UserRepository) ByEmail(_ context.Context, tenantID tenantmodel.TenantID, email usermodel.Email) (usermodel.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.idByTenantEmail[tenantEmailKey(tenantID, email)]
	if !ok {
		return usermodel.User{}, fmt.Errorf("%w: %s/%s", usermodel.ErrUserNotFound, tenantID.String(), email.String())
	}
	return r.byID[id], nil
}

func (r *UserRepository) CountByTenant(_ context.Context, tenantID tenantmodel.TenantID) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, user := range r.byID {
		if user.TenantID().Equal(tenantID) {
			count++
		}
	}
	return count, nil
}
