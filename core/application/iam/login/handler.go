package login

import (
	"context"
	"errors"
	"fmt"

	"github.com/usegavel/gavel/core/application/shared/event"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

type Handler struct {
	tenants  service.TenantRepository
	users    service.UserRepository
	sessions service.SessionRepository
	hasher   service.PasswordHasher
	secrets  service.SecretGenerator
}

func NewHandler(tenants service.TenantRepository, users service.UserRepository, sessions service.SessionRepository, hasher service.PasswordHasher, secrets service.SecretGenerator) *Handler {
	if tenants == nil {
		panic("iam/login: tenants repository must not be nil")
	}
	if users == nil {
		panic("iam/login: users repository must not be nil")
	}
	if sessions == nil {
		panic("iam/login: sessions repository must not be nil")
	}
	if hasher == nil {
		panic("iam/login: password hasher must not be nil")
	}
	if secrets == nil {
		panic("iam/login: secret generator must not be nil")
	}
	return &Handler{tenants: tenants, users: users, sessions: sessions, hasher: hasher, secrets: secrets}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	slug, err := tenant.NewSlug(cmd.TenantSlug())
	if err != nil {
		return Result{}, ErrInvalidCredentials
	}
	email, err := user.NewEmail(cmd.Email())
	if err != nil {
		return Result{}, ErrInvalidCredentials
	}

	foundTenant, err := h.tenants.BySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) {
			return Result{}, ErrInvalidCredentials
		}
		return Result{}, fmt.Errorf("find tenant: %w", err)
	}
	if !foundTenant.Status().IsActive() {
		return Result{}, ErrInvalidCredentials
	}

	foundUser, err := h.users.ByEmail(ctx, foundTenant.ID(), email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return Result{}, ErrInvalidCredentials
		}
		return Result{}, fmt.Errorf("find user: %w", err)
	}
	if !foundUser.IsActive() {
		return Result{}, ErrInvalidCredentials
	}

	ok, err := h.hasher.Verify(cmd.PlainPassword(), foundUser.PasswordHash())
	if err != nil {
		return Result{}, fmt.Errorf("verify password: %w", err)
	}
	if !ok {
		return Result{}, ErrInvalidCredentials
	}

	sessionToken, err := h.secrets.NewSessionToken()
	if err != nil {
		return Result{}, fmt.Errorf("generate session token: %w", err)
	}

	expiresAt := cmd.OccurredAt().Add(cmd.SessionTTL())
	sess, err := session.NewSession(sessionToken, foundUser.ID(), cmd.UserAgent(), cmd.IPAddress(), cmd.OccurredAt(), expiresAt)
	if err != nil {
		return Result{}, fmt.Errorf("new session: %w", err)
	}

	events := sess.Events()
	sess.ClearEvents()

	if err := h.sessions.Save(ctx, sess); err != nil {
		return Result{}, fmt.Errorf("save session: %w", err)
	}

	if err := foundUser.TouchLogin(cmd.OccurredAt()); err != nil {
		return Result{}, fmt.Errorf("touch login: %w", err)
	}
	if err := h.users.Save(ctx, foundUser); err != nil {
		return Result{}, fmt.Errorf("save user: %w", err)
	}

	return Result{
		UserID:             foundUser.ID().String(),
		TenantID:           foundTenant.ID().String(),
		Email:              foundUser.Email().String(),
		DisplayName:        foundUser.DisplayName(),
		Role:               foundUser.Role().String(),
		MustChangePassword: foundUser.MustChangePassword(),
		SessionToken:       sessionToken.String(),
		SessionExpiresAt:   expiresAt,
		Events:             event.EventsFromDomain(events),
	}, nil
}
