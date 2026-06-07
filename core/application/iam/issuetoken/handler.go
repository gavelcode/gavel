package issuetoken

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/application/shared/event"
	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

type Handler struct {
	users   service.UserRepository
	tokens  service.APITokenRepository
	secrets service.SecretGenerator
}

func NewHandler(users service.UserRepository, tokens service.APITokenRepository, secrets service.SecretGenerator) *Handler {
	if users == nil {
		panic("iam/issuetoken: users repository must not be nil")
	}
	if tokens == nil {
		panic("iam/issuetoken: tokens repository must not be nil")
	}
	if secrets == nil {
		panic("iam/issuetoken: secret generator must not be nil")
	}
	return &Handler{users: users, tokens: tokens, secrets: secrets}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	userID, err := user.ParseUserID(cmd.UserID())
	if err != nil {
		return Result{}, err
	}
	scopes, err := parseScopes(cmd.Scopes())
	if err != nil {
		return Result{}, err
	}
	foundUser, err := h.users.ByID(ctx, userID)
	if err != nil {
		return Result{}, fmt.Errorf("find user: %w", err)
	}
	if !foundUser.IsActive() {
		return Result{}, fmt.Errorf("%w: cannot issue token for inactive user", apitoken.ErrInvalid)
	}

	secret, err := h.secrets.NewAPITokenSecret()
	if err != nil {
		return Result{}, fmt.Errorf("generate api token secret: %w", err)
	}

	token, err := apitoken.NewAPIToken(secret, foundUser.TenantID(), foundUser.ID(), cmd.Name(), scopes, cmd.OccurredAt(), cmd.ExpiresAt())
	if err != nil {
		return Result{}, fmt.Errorf("new api token: %w", err)
	}

	events := token.Events()
	token.ClearEvents()

	if err := h.tokens.Save(ctx, token); err != nil {
		return Result{}, fmt.Errorf("save api token: %w", err)
	}

	tokenScopes := token.Scopes()
	scopeStrings := make([]string, 0, len(tokenScopes))
	for _, s := range tokenScopes {
		scopeStrings = append(scopeStrings, s.String())
	}

	return Result{
		TokenID:     token.ID().String(),
		UserID:      token.UserID().String(),
		TenantID:    token.TenantID().String(),
		Name:        token.Name(),
		TokenPrefix: token.TokenPrefix(),
		Scopes:      scopeStrings,
		PlainSecret: secret.String(),
		Events:      event.EventsFromDomain(events),
	}, nil
}

func parseScopes(raw []string) (apitoken.Scopes, error) {
	out := make(apitoken.Scopes, 0, len(raw))
	for _, r := range raw {
		s, err := apitoken.NewScope(r)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}
