package logout

import (
	"context"
	"fmt"

	"github.com/usegavel/gavel/core/application/shared/event"
	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/service"
)

type Handler struct {
	sessions service.SessionRepository
}

func NewHandler(sessions service.SessionRepository) *Handler {
	if sessions == nil {
		panic("iam/logout: sessions repository must not be nil")
	}
	return &Handler{sessions: sessions}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (Result, error) {
	token, err := session.NewToken(cmd.SessionToken())
	if err != nil {
		return Result{}, fmt.Errorf("parse session token: %w", err)
	}
	hash := session.HashToken(token)

	sess, err := h.sessions.ByTokenHash(ctx, hash)
	if err != nil {
		return Result{}, fmt.Errorf("find session: %w", err)
	}

	if err := sess.Revoke(cmd.OccurredAt()); err != nil {
		return Result{}, fmt.Errorf("revoke session: %w", err)
	}

	events := sess.Events()
	sess.ClearEvents()

	if err := h.sessions.Save(ctx, sess); err != nil {
		return Result{}, fmt.Errorf("save session: %w", err)
	}

	return Result{
		SessionID: sess.ID().String(),
		Events:    event.EventsFromDomain(events),
	}, nil
}
