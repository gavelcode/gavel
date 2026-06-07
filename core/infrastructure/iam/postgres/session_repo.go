package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/usegavel/gavel/core/domain/iam/model/session"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

type SessionRepo struct {
	db *database.DB
}

var _ service.SessionRepository = (*SessionRepo)(nil)

func NewSessionRepo(db *database.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Save(ctx context.Context, sess session.Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO iam_sessions (
			id, token_hash, user_id, user_agent, ip_address,
			created_at, expires_at, last_seen_at, is_revoked
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			last_seen_at = EXCLUDED.last_seen_at,
			is_revoked = EXCLUDED.is_revoked,
			expires_at = EXCLUDED.expires_at
	`,
		sess.ID().UUID(), sess.TokenHash().String(), sess.UserID().UUID(),
		sess.UserAgent(), sess.IPAddress(),
		sess.CreatedAt(), sess.ExpiresAt(), sess.LastSeenAt(),
		sess.IsRevoked(),
	)
	if err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	return nil
}

func (r *SessionRepo) ByTokenHash(ctx context.Context, hash session.TokenHash) (session.Session, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, token_hash, user_id, user_agent, ip_address,
		       created_at, expires_at, last_seen_at, is_revoked
		FROM iam_sessions
		WHERE token_hash = ?
	`, hash.String())

	var idVal, userIDVal uuid.UUID
	var tokenHashRaw, userAgent, ipAddress string
	var createdAt, expiresAt, lastSeenAt sql.NullTime
	var isRevoked bool
	if err := row.Scan(&idVal, &tokenHashRaw, &userIDVal, &userAgent, &ipAddress, &createdAt, &expiresAt, &lastSeenAt, &isRevoked); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return session.Session{}, fmt.Errorf("%w", session.ErrNotFound)
		}
		return session.Session{}, fmt.Errorf("scan session: %w", err)
	}

	tokenHash, err := session.NewTokenHash(tokenHashRaw)
	if err != nil {
		return session.Session{}, fmt.Errorf("hydrate token hash: %w", err)
	}
	id := session.NewSessionID(idVal)
	userID := user.NewUserID(userIDVal)
	return session.ReconstituteSession(id, tokenHash, userID, userAgent, ipAddress, createdAt.Time, expiresAt.Time, lastSeenAt.Time, isRevoked)
}

func (r *SessionRepo) DeleteAllForUser(ctx context.Context, userID user.UserID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM iam_sessions WHERE user_id = ?`, userID.UUID())
	if err != nil {
		return fmt.Errorf("delete sessions for user: %w", err)
	}
	return nil
}

func (r *SessionRepo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM iam_sessions WHERE is_revoked = true OR expires_at <= ?`, before)
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}
