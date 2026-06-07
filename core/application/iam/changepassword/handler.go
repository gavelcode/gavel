package changepassword

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
	hasher   service.PasswordHasher
}

func NewHandler(users service.UserRepository, sessions service.SessionRepository, hasher service.PasswordHasher) *Handler {
	if users == nil {
		panic("iam/changepassword: users repository must not be nil")
	}
	if sessions == nil {
		panic("iam/changepassword: sessions repository must not be nil")
	}
	if hasher == nil {
		panic("iam/changepassword: password hasher must not be nil")
	}
	return &Handler{users: users, sessions: sessions, hasher: hasher}
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

	ok, err := h.hasher.Verify(cmd.CurrentPassword(), foundUser.PasswordHash())
	if err != nil {
		return Result{}, fmt.Errorf("verify password: %w", err)
	}
	if !ok {
		return Result{}, ErrCurrentPasswordWrong
	}

	newHash, err := h.hasher.Hash(cmd.NewPassword())
	if err != nil {
		return Result{}, fmt.Errorf("hash password: %w", err)
	}
	if err := foundUser.ChangePassword(newHash, cmd.OccurredAt()); err != nil {
		return Result{}, fmt.Errorf("change password: %w", err)
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
