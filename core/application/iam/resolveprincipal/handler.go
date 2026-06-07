package resolveprincipal

import (
	"context"
	"errors"
	"fmt"

	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

type Handler struct {
	users    service.UserRepository
	sessions service.SessionRepository
	tokens   service.APITokenRepository
}

func NewHandler(users service.UserRepository, sessions service.SessionRepository, tokens service.APITokenRepository) *Handler {
	if users == nil {
		panic("iam/resolveprincipal: users repository must not be nil")
	}
	if sessions == nil {
		panic("iam/resolveprincipal: sessions repository must not be nil")
	}
	if tokens == nil {
		panic("iam/resolveprincipal: tokens repository must not be nil")
	}
	return &Handler{users: users, sessions: sessions, tokens: tokens}
}

func (h *Handler) Execute(ctx context.Context, query Query) (Principal, error) {
	if query.BearerToken() != "" {
		return h.resolveBearer(ctx, query)
	}
	if query.SessionCookie() != "" {
		return h.resolveCookie(ctx, query)
	}
	return Principal{}, ErrUnauthenticated
}

func (h *Handler) resolveBearer(ctx context.Context, query Query) (Principal, error) {
	secret, err := apitoken.NewSecret(query.BearerToken())
	if err != nil {
		return Principal{}, ErrUnauthenticated
	}
	hash := apitoken.HashSecret(secret)

	token, err := h.tokens.ByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, apitoken.ErrNotFound) {
			return Principal{}, ErrUnauthenticated
		}
		return Principal{}, fmt.Errorf("find token: %w", err)
	}
	if token.IsExpired(query.OccurredAt()) {
		return Principal{}, ErrUnauthenticated
	}

	foundUser, err := h.users.ByID(ctx, token.UserID())
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return Principal{}, ErrUnauthenticated
		}
		return Principal{}, fmt.Errorf("find token user: %w", err)
	}
	if !foundUser.IsActive() {
		return Principal{}, ErrUnauthenticated
	}

	scopes := token.Scopes()
	scopeStrings := make([]string, 0, len(scopes))
	for _, s := range scopes {
		scopeStrings = append(scopeStrings, s.String())
	}

	return Principal{
		UserID:             foundUser.ID().String(),
		TenantID:           foundUser.TenantID().String(),
		Email:              foundUser.Email().String(),
		DisplayName:        foundUser.DisplayName(),
		Role:               foundUser.Role().String(),
		MustChangePassword: foundUser.MustChangePassword(),
		ViaAPIToken:        true,
		APITokenID:         token.ID().String(),
		Scopes:             scopeStrings,
	}, nil
}

func (h *Handler) resolveCookie(ctx context.Context, query Query) (Principal, error) {
	cookie, err := session.NewToken(query.SessionCookie())
	if err != nil {
		return Principal{}, ErrUnauthenticated
	}
	hash := session.HashToken(cookie)

	sess, err := h.sessions.ByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, session.ErrNotFound) {
			return Principal{}, ErrUnauthenticated
		}
		return Principal{}, fmt.Errorf("find session: %w", err)
	}
	if sess.IsExpired(query.OccurredAt()) {
		return Principal{}, ErrUnauthenticated
	}

	foundUser, err := h.users.ByID(ctx, sess.UserID())
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return Principal{}, ErrUnauthenticated
		}
		return Principal{}, fmt.Errorf("find session user: %w", err)
	}
	if !foundUser.IsActive() {
		return Principal{}, ErrUnauthenticated
	}

	return Principal{
		UserID:             foundUser.ID().String(),
		TenantID:           foundUser.TenantID().String(),
		Email:              foundUser.Email().String(),
		DisplayName:        foundUser.DisplayName(),
		Role:               foundUser.Role().String(),
		MustChangePassword: foundUser.MustChangePassword(),
		ViaAPIToken:        false,
	}, nil
}
