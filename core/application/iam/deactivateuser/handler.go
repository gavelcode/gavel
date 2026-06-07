package deactivateuser

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/application/shared/event"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

type Handler struct {
	users    service.UserRepository
	sessions service.SessionRepository
}

func NewHandler(users service.UserRepository, sessions service.SessionRepository) *Handler {
	if users == nil {
		panic("iam/deactivateuser: users repository must not be nil")
	}
	if sessions == nil {
		panic("iam/deactivateuser: sessions repository must not be nil")
	}
	return &Handler{users: users, sessions: sessions}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	userID, err := user.ParseUserID(cmd.UserID())
	if err != nil {
		return Result{}, err
	}
	foundUser, err := h.users.ByID(ctx, userID)
	if err != nil {
		return Result{}, fmt.Errorf("find user: %w", err)
	}
	if err := foundUser.Deactivate(cmd.OccurredAt()); err != nil {
		return Result{}, fmt.Errorf("deactivate user: %w", err)
	}

	events := foundUser.Events()
	foundUser.ClearEvents()

	if err := h.users.Save(ctx, foundUser); err != nil {
		return Result{}, fmt.Errorf("save user: %w", err)
	}
	if err := h.sessions.DeleteAllForUser(ctx, foundUser.ID()); err != nil {
		return Result{}, fmt.Errorf("delete sessions: %w", err)
	}

	return Result{
		UserID: foundUser.ID().String(),
		Events: event.EventsFromDomain(events),
	}, nil
}
