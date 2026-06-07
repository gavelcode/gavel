package listmytokens

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

type Handler struct {
	tokens service.APITokenRepository
}

func NewHandler(tokens service.APITokenRepository) *Handler {
	if tokens == nil {
		panic("iam/listmytokens: tokens repository must not be nil")
	}
	return &Handler{tokens: tokens}
}

func (h *Handler) Execute(ctx context.Context, query Query) (Result, error) {
	userID, err := user.ParseUserID(query.UserID())
	if err != nil {
		return Result{}, err
	}
	tokens, err := h.tokens.ListByUser(ctx, userID)
	if err != nil {
		return Result{}, fmt.Errorf("list tokens: %w", err)
	}

	now := query.Now()
	out := make([]TokenSummary, 0, len(tokens))
	for _, token := range tokens {
		scopes := make([]string, 0, len(token.Scopes()))
		for _, s := range token.Scopes() {
			scopes = append(scopes, s.String())
		}
		summary := TokenSummary{
			ID:        token.ID().String(),
			Name:      token.Name(),
			Prefix:    token.TokenPrefix(),
			Scopes:    scopes,
			CreatedAt: token.CreatedAt(),
			IsRevoked: token.IsRevoked(),
			IsExpired: token.IsExpired(now),
		}
		if !token.LastUsedAt().IsZero() {
			lu := token.LastUsedAt()
			summary.LastUsedAt = &lu
		}
		if !token.ExpiresAt().IsZero() {
			ex := token.ExpiresAt()
			summary.ExpiresAt = &ex
		}
		out = append(out, summary)
	}
	return Result{Tokens: out}, nil
}
