package user

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/shared/event"
)

type User struct {
	id                 UserID
	tenantID           tenant.TenantID
	email              Email
	displayName        string
	role               Role
	passwordHash       PasswordHash
	mustChangePassword bool
	isActive           bool
	createdAt          time.Time
	lastLoginAt        time.Time
	events             []event.DomainEvent
}

func NewUser(
	tenantID tenant.TenantID,
	email Email,
	displayName string,
	role Role,
	passwordHash PasswordHash,
	mustChangePassword bool,
	createdAt time.Time,
) (User, error) {
	if err := validateUserFields(tenantID, email, displayName, role, passwordHash, createdAt); err != nil {
		return User{}, err
	}
	userID := NewUserID(uuid.New())
	trimmedName := strings.TrimSpace(displayName)
	return User{
		id:                 userID,
		tenantID:           tenantID,
		email:              email,
		displayName:        trimmedName,
		role:               role,
		passwordHash:       passwordHash,
		mustChangePassword: mustChangePassword,
		isActive:           true,
		createdAt:          createdAt,
		events:             []event.DomainEvent{NewUserCreated(userID, tenantID, email, role, mustChangePassword, createdAt)},
	}, nil
}

func ReconstituteUser(
	userID UserID,
	tenantID tenant.TenantID,
	email Email,
	displayName string,
	role Role,
	passwordHash PasswordHash,
	mustChangePassword bool,
	isActive bool,
	createdAt time.Time,
	lastLoginAt time.Time,
) (User, error) {
	if err := validateUserFields(tenantID, email, displayName, role, passwordHash, createdAt); err != nil {
		return User{}, err
	}
	return User{
		id:                 userID,
		tenantID:           tenantID,
		email:              email,
		displayName:        strings.TrimSpace(displayName),
		role:               role,
		passwordHash:       passwordHash,
		mustChangePassword: mustChangePassword,
		isActive:           isActive,
		createdAt:          createdAt,
		lastLoginAt:        lastLoginAt,
	}, nil
}

func validateUserFields(tenantID tenant.TenantID, email Email, displayName string, role Role, passwordHash PasswordHash, createdAt time.Time) error {
	if strings.TrimSpace(displayName) == "" {
		return fmt.Errorf("%w: display name must not be empty", ErrInvalidUser)
	}
	if createdAt.IsZero() {
		return fmt.Errorf("%w: createdAt must not be zero", ErrInvalidUser)
	}
	return nil
}

func (u *User) ID() UserID                 { return u.id }
func (u *User) TenantID() tenant.TenantID  { return u.tenantID }
func (u *User) Email() Email               { return u.email }
func (u *User) DisplayName() string        { return u.displayName }
func (u *User) Role() Role                 { return u.role }
func (u *User) PasswordHash() PasswordHash { return u.passwordHash }
func (u *User) MustChangePassword() bool   { return u.mustChangePassword }
func (u *User) IsActive() bool             { return u.isActive }
func (u *User) CreatedAt() time.Time       { return u.createdAt }
func (u *User) LastLoginAt() time.Time     { return u.lastLoginAt }

func (u *User) ChangePassword(newHash PasswordHash, occurredAt time.Time) error {
	if occurredAt.IsZero() {
		return fmt.Errorf("%w: occurredAt must not be zero", ErrInvalidUser)
	}
	u.passwordHash = newHash
	u.mustChangePassword = false
	u.events = append(u.events, NewPasswordChanged(u.id, occurredAt))
	return nil
}

func (u *User) Deactivate(occurredAt time.Time) error {
	if !u.isActive {
		return fmt.Errorf("%w: user is already inactive", ErrInvalidUser)
	}
	if occurredAt.IsZero() {
		return fmt.Errorf("%w: occurredAt must not be zero", ErrInvalidUser)
	}
	u.isActive = false
	u.events = append(u.events, NewUserDeactivated(u.id, occurredAt))
	return nil
}

func (u *User) TouchLogin(at time.Time) error {
	if at.IsZero() {
		return fmt.Errorf("%w: login timestamp must not be zero", ErrInvalidUser)
	}
	u.lastLoginAt = at
	return nil
}

func (u *User) Events() []event.DomainEvent {
	copied := make([]event.DomainEvent, len(u.events))
	copy(copied, u.events)
	return copied
}

func (u *User) ClearEvents() {
	u.events = nil
}
