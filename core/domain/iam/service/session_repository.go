package service

import (
	"context"
	"time"

	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
)

type SessionRepository interface {
	Save(ctx context.Context, sess session.Session) error
	ByTokenHash(ctx context.Context, hash session.TokenHash) (session.Session, error)
	DeleteAllForUser(ctx context.Context, userID user.UserID) error
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}
