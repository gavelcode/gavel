package tenant

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/usegavel/gavel/core/domain/shared/event"
)

type Tenant struct {
	id          TenantID
	slug        Slug
	displayName string
	status      Status
	createdAt   time.Time
	events      []event.DomainEvent
}

func NewTenant(slug Slug, displayName string, createdAt time.Time) (Tenant, error) {
	trimmedName := strings.TrimSpace(displayName)
	if trimmedName == "" {
		return Tenant{}, fmt.Errorf("%w: display name must not be empty", ErrInvalidTenant)
	}
	if createdAt.IsZero() {
		return Tenant{}, fmt.Errorf("%w: createdAt must not be zero", ErrInvalidTenant)
	}
	tenantID := NewTenantID(uuid.New())
	return Tenant{
		id:          tenantID,
		slug:        slug,
		displayName: trimmedName,
		status:      StatusActive,
		createdAt:   createdAt,
		events:      []event.DomainEvent{NewTenantCreated(tenantID, slug, trimmedName, createdAt)},
	}, nil
}

func ReconstituteTenant(tenantID TenantID, slug Slug, displayName string, status Status, createdAt time.Time) (Tenant, error) {
	if strings.TrimSpace(displayName) == "" {
		return Tenant{}, fmt.Errorf("%w: display name must not be empty", ErrInvalidTenant)
	}
	if createdAt.IsZero() {
		return Tenant{}, fmt.Errorf("%w: createdAt must not be zero", ErrInvalidTenant)
	}
	return Tenant{
		id:          tenantID,
		slug:        slug,
		displayName: displayName,
		status:      status,
		createdAt:   createdAt,
	}, nil
}

func (t *Tenant) ID() TenantID         { return t.id }
func (t *Tenant) Slug() Slug           { return t.slug }
func (t *Tenant) DisplayName() string  { return t.displayName }
func (t *Tenant) Status() Status       { return t.status }
func (t *Tenant) CreatedAt() time.Time { return t.createdAt }

func (t *Tenant) Suspend(occurredAt time.Time) error {
	if !t.status.IsActive() {
		return fmt.Errorf("%w: tenant is not active", ErrInvalidTenant)
	}
	if occurredAt.IsZero() {
		return fmt.Errorf("%w: occurredAt must not be zero", ErrInvalidTenant)
	}
	t.status = StatusSuspended
	t.events = append(t.events, NewTenantSuspended(t.id, occurredAt))
	return nil
}

func (t *Tenant) Activate(occurredAt time.Time) error {
	if t.status.IsActive() {
		return fmt.Errorf("%w: tenant is already active", ErrInvalidTenant)
	}
	if occurredAt.IsZero() {
		return fmt.Errorf("%w: occurredAt must not be zero", ErrInvalidTenant)
	}
	t.status = StatusActive
	t.events = append(t.events, NewTenantActivated(t.id, occurredAt))
	return nil
}

func (t *Tenant) Events() []event.DomainEvent {
	copied := make([]event.DomainEvent, len(t.events))
	copy(copied, t.events)
	return copied
}

func (t *Tenant) ClearEvents() {
	t.events = nil
}
