package user

import (
	"time"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

const (
	EventNameUserCreated     = "iam.user_created"
	EventNamePasswordChanged = "iam.password_changed"
	EventNameUserDeactivated = "iam.user_deactivated"
)

type UserCreated struct {
	userID             UserID
	tenantID           tenant.TenantID
	email              Email
	role               Role
	mustChangePassword bool
	occurredAt         time.Time
}

func NewUserCreated(userID UserID, tenantID tenant.TenantID, email Email, role Role, mustChangePassword bool, occurredAt time.Time) UserCreated {
	return UserCreated{
		userID:             userID,
		tenantID:           tenantID,
		email:              email,
		role:               role,
		mustChangePassword: mustChangePassword,
		occurredAt:         occurredAt,
	}
}

func (e UserCreated) EventName() string         { return EventNameUserCreated }
func (e UserCreated) OccurredAt() time.Time     { return e.occurredAt }
func (e UserCreated) UserID() UserID            { return e.userID }
func (e UserCreated) TenantID() tenant.TenantID { return e.tenantID }
func (e UserCreated) Email() Email              { return e.email }
func (e UserCreated) Role() Role                { return e.role }
func (e UserCreated) MustChangePassword() bool  { return e.mustChangePassword }

type PasswordChanged struct {
	userID     UserID
	occurredAt time.Time
}

func NewPasswordChanged(userID UserID, occurredAt time.Time) PasswordChanged {
	return PasswordChanged{userID: userID, occurredAt: occurredAt}
}

func (e PasswordChanged) EventName() string     { return EventNamePasswordChanged }
func (e PasswordChanged) OccurredAt() time.Time { return e.occurredAt }
func (e PasswordChanged) UserID() UserID        { return e.userID }

type UserDeactivated struct {
	userID     UserID
	occurredAt time.Time
}

func NewUserDeactivated(userID UserID, occurredAt time.Time) UserDeactivated {
	return UserDeactivated{userID: userID, occurredAt: occurredAt}
}

func (e UserDeactivated) EventName() string     { return EventNameUserDeactivated }
func (e UserDeactivated) OccurredAt() time.Time { return e.occurredAt }
func (e UserDeactivated) UserID() UserID        { return e.userID }
