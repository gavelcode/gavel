package revoketoken

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/application/shared/event"
	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

type Handler struct {
	tokens service.APITokenRepository
}

func NewHandler(tokens service.APITokenRepository) *Handler {
	if tokens == nil {
		panic("iam/revoketoken: tokens repository must not be nil")
	}
	return &Handler{tokens: tokens}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	tokenID, err := apitoken.ParseAPITokenID(cmd.TokenID())
	if err != nil {
		return Result{}, err
	}
	callerID, err := user.ParseUserID(cmd.CallerUserID())
	if err != nil {
		return Result{}, err
	}
	token, err := h.tokens.ByID(ctx, tokenID)
	if err != nil {
		return Result{}, fmt.Errorf("find token: %w", err)
	}
	if !token.UserID().Equal(callerID) {
		return Result{}, fmt.Errorf("%w: token belongs to %s, caller is %s", ErrUnauthorized, token.UserID(), callerID)
	}
	if err := token.Revoke(cmd.OccurredAt()); err != nil {
		return Result{}, fmt.Errorf("revoke token: %w", err)
	}

	events := token.Events()
	token.ClearEvents()

	if err := h.tokens.Save(ctx, token); err != nil {
		return Result{}, fmt.Errorf("save token: %w", err)
	}

	return Result{
		TokenID: token.ID().String(),
		Events:  event.EventsFromDomain(events),
	}, nil
}
