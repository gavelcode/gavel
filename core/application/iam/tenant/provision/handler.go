package provision

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/application/shared/event"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

const adminMustChangePassword = true

type Handler struct {
	tenants service.TenantRepository
	users   service.UserRepository
	hasher  service.PasswordHasher
}

func NewHandler(tenants service.TenantRepository, users service.UserRepository, hasher service.PasswordHasher) *Handler {
	if tenants == nil {
		panic("iam/tenant/provision: tenants repository must not be nil")
	}
	if users == nil {
		panic("iam/tenant/provision: users repository must not be nil")
	}
	if hasher == nil {
		panic("iam/tenant/provision: password hasher must not be nil")
	}
	return &Handler{tenants: tenants, users: users, hasher: hasher}
}

// Execute provisions a tenant together with its first administrator, mirroring
// Vernon's coarse provisionTenant: a tenant without an admin is unusable, so the
// two are created as one use case. The admin is forced to change the password on
// first login. The two saves are not one transaction — a failure after the
// tenant is saved leaves a tenant without an admin; the operator re-runs (the
// slug is now taken, so the retry is rejected) or provisions the admin directly.
func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	slug, err := tenant.NewSlug(cmd.Slug())
	if err != nil {
		return Result{}, err
	}
	email, err := user.NewEmail(cmd.AdminEmail())
	if err != nil {
		return Result{}, err
	}
	newTenant, err := tenant.NewTenant(slug, cmd.DisplayName(), cmd.OccurredAt())
	if err != nil {
		return Result{}, fmt.Errorf("new tenant: %w", err)
	}
	hash, err := h.hasher.Hash(cmd.AdminPassword())
	if err != nil {
		return Result{}, fmt.Errorf("hash admin password: %w", err)
	}
	adminUser, err := user.NewUser(
		newTenant.ID(), email, cmd.AdminDisplayName(), user.RoleAdmin, hash, adminMustChangePassword, cmd.OccurredAt())
	if err != nil {
		return Result{}, fmt.Errorf("new admin user: %w", err)
	}

	tenantEvents := newTenant.Events()
	newTenant.ClearEvents()
	if err := h.tenants.Save(ctx, newTenant); err != nil {
		return Result{}, fmt.Errorf("save tenant: %w", err)
	}

	userEvents := adminUser.Events()
	adminUser.ClearEvents()
	if err := h.users.Save(ctx, adminUser); err != nil {
		return Result{}, fmt.Errorf("save admin user: %w", err)
	}

	events := append(event.EventsFromDomain(tenantEvents), event.EventsFromDomain(userEvents)...)
	return Result{
		TenantID:    newTenant.ID().String(),
		AdminUserID: adminUser.ID().String(),
		Events:      events,
	}, nil
}
