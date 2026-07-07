package tenant

import "time"

const (
	EventNameTenantCreated   = "iam.tenant_created"
	EventNameTenantSuspended = "iam.tenant_suspended"
	EventNameTenantActivated = "iam.tenant_activated"
)

type TenantCreated struct {
	tenantID    TenantID
	slug        Slug
	displayName string
	occurredAt  time.Time
}

func NewTenantCreated(tenantID TenantID, slug Slug, displayName string, occurredAt time.Time) TenantCreated {
	return TenantCreated{
		tenantID:    tenantID,
		slug:        slug,
		displayName: displayName,
		occurredAt:  occurredAt,
	}
}

func (e TenantCreated) EventName() string     { return EventNameTenantCreated }
func (e TenantCreated) OccurredAt() time.Time { return e.occurredAt }
func (e TenantCreated) TenantID() TenantID    { return e.tenantID }
func (e TenantCreated) Slug() Slug            { return e.slug }
func (e TenantCreated) DisplayName() string   { return e.displayName }

type TenantSuspended struct {
	tenantID   TenantID
	occurredAt time.Time
}

func NewTenantSuspended(tenantID TenantID, occurredAt time.Time) TenantSuspended {
	return TenantSuspended{tenantID: tenantID, occurredAt: occurredAt}
}

func (e TenantSuspended) EventName() string     { return EventNameTenantSuspended }
func (e TenantSuspended) OccurredAt() time.Time { return e.occurredAt }
func (e TenantSuspended) TenantID() TenantID    { return e.tenantID }

type TenantActivated struct {
	tenantID   TenantID
	occurredAt time.Time
}

func NewTenantActivated(tenantID TenantID, occurredAt time.Time) TenantActivated {
	return TenantActivated{tenantID: tenantID, occurredAt: occurredAt}
}

func (e TenantActivated) EventName() string     { return EventNameTenantActivated }
func (e TenantActivated) OccurredAt() time.Time { return e.occurredAt }
func (e TenantActivated) TenantID() TenantID    { return e.tenantID }
