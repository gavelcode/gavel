package createuser

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/application/shared/event"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

type Handler struct {
	tenants service.TenantRepository
	users   service.UserRepository
	hasher  service.PasswordHasher
}

func NewHandler(tenants service.TenantRepository, users service.UserRepository, hasher service.PasswordHasher) *Handler {
	if tenants == nil {
		panic("iam/createuser: tenants repository must not be nil")
	}
	if users == nil {
		panic("iam/createuser: users repository must not be nil")
	}
	if hasher == nil {
		panic("iam/createuser: password hasher must not be nil")
	}
	return &Handler{tenants: tenants, users: users, hasher: hasher}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	tenantID, err := tenant.ParseTenantID(cmd.TenantID())
	if err != nil {
		return Result{}, err
	}
	email, err := user.NewEmail(cmd.Email())
	if err != nil {
		return Result{}, err
	}
	role, err := user.NewRole(cmd.Role())
	if err != nil {
		return Result{}, err
	}
	foundTenant, err := h.tenants.ByID(ctx, tenantID)
	if err != nil {
		return Result{}, fmt.Errorf("find tenant: %w", err)
	}
	if !foundTenant.Status().IsActive() {
		return Result{}, fmt.Errorf("%w: cannot create user under a suspended tenant", user.ErrInvalidUser)
	}

	hash, err := h.hasher.Hash(cmd.PlainPassword())
	if err != nil {
		return Result{}, fmt.Errorf("hash password: %w", err)
	}

	newUser, err := user.NewUser(tenantID, email, cmd.DisplayName(), role, hash, cmd.MustChangePassword(), cmd.OccurredAt())
	if err != nil {
		return Result{}, fmt.Errorf("new user: %w", err)
	}

	events := newUser.Events()
	newUser.ClearEvents()

	if err := h.users.Save(ctx, newUser); err != nil {
		return Result{}, fmt.Errorf("save user: %w", err)
	}

	return Result{
		UserID:      newUser.ID().String(),
		TenantID:    newUser.TenantID().String(),
		Email:       newUser.Email().String(),
		DisplayName: newUser.DisplayName(),
		Role:        newUser.Role().String(),
		Events:      event.EventsFromDomain(events),
	}, nil
}
