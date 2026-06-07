package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/usegavel/gavel/core/domain/iam/model/apitoken"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

type APITokenRepo struct {
	db *database.DB
}

var _ service.APITokenRepository = (*APITokenRepo)(nil)

func NewAPITokenRepo(db *database.DB) *APITokenRepo {
	return &APITokenRepo{db: db}
}

func (r *APITokenRepo) Save(ctx context.Context, token apitoken.APIToken) error {
	scopes := token.Scopes()
	scopeStrings := make([]string, 0, len(scopes))
	for _, s := range scopes {
		scopeStrings = append(scopeStrings, s.String())
	}
	scopesJSON, err := json.Marshal(scopeStrings)
	if err != nil {
		return fmt.Errorf("marshal scopes: %w", err)
	}

	var expiresAt, lastUsedAt any
	if !token.ExpiresAt().IsZero() {
		expiresAt = token.ExpiresAt()
	}
	if !token.LastUsedAt().IsZero() {
		lastUsedAt = token.LastUsedAt()
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO iam_api_tokens (
			id, tenant_id, user_id, name, token_hash, token_prefix,
			scopes, created_at, expires_at, last_used_at, is_revoked
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			scopes = EXCLUDED.scopes,
			expires_at = EXCLUDED.expires_at,
			last_used_at = EXCLUDED.last_used_at,
			is_revoked = EXCLUDED.is_revoked
	`,
		token.ID().UUID(), token.TenantID().UUID(), token.UserID().UUID(),
		token.Name(), token.TokenHash().String(), token.TokenPrefix(),
		string(scopesJSON), token.CreatedAt(), expiresAt, lastUsedAt, token.IsRevoked(),
	)
	if err != nil {
		return fmt.Errorf("save api token: %w", err)
	}
	return nil
}

func (r *APITokenRepo) ByID(ctx context.Context, id apitoken.APITokenID) (apitoken.APIToken, error) {
	return r.scanOne(ctx, `WHERE id = ?`, id.UUID())
}

func (r *APITokenRepo) ByTokenHash(ctx context.Context, hash apitoken.SecretHash) (apitoken.APIToken, error) {
	return r.scanOne(ctx, `WHERE token_hash = ?`, hash.String())
}

func (r *APITokenRepo) ListByUser(ctx context.Context, userID user.UserID) ([]apitoken.APIToken, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, user_id, name, token_hash, token_prefix,
		       scopes, created_at, expires_at, last_used_at, is_revoked
		FROM iam_api_tokens
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID.UUID())
	if err != nil {
		return nil, fmt.Errorf("list tokens by user: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []apitoken.APIToken
	for rows.Next() {
		tok, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, tok)
	}
	return out, rows.Err()
}

func (r *APITokenRepo) scanOne(ctx context.Context, where string, args ...any) (apitoken.APIToken, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, user_id, name, token_hash, token_prefix,
		       scopes, created_at, expires_at, last_used_at, is_revoked
		FROM iam_api_tokens
	`+where, args...)
	tok, err := r.scanRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apitoken.APIToken{}, fmt.Errorf("%w", apitoken.ErrNotFound)
		}
		return apitoken.APIToken{}, err
	}
	return tok, nil
}

func (r *APITokenRepo) scanRow(row scanner) (apitoken.APIToken, error) {
	var idVal, tenantIDVal, userIDVal uuid.UUID
	var name, hashRaw, prefix, scopesRaw string
	var createdAt sql.NullTime
	var expiresAt, lastUsedAt sql.NullTime
	var isRevoked bool
	if err := row.Scan(&idVal, &tenantIDVal, &userIDVal, &name, &hashRaw, &prefix, &scopesRaw, &createdAt, &expiresAt, &lastUsedAt, &isRevoked); err != nil {
		return apitoken.APIToken{}, err
	}

	tokenID := apitoken.NewAPITokenID(idVal)
	tenantID := tenant.NewTenantID(tenantIDVal)
	userID := user.NewUserID(userIDVal)
	hash, err := apitoken.NewSecretHash(hashRaw)
	if err != nil {
		return apitoken.APIToken{}, fmt.Errorf("hydrate token hash: %w", err)
	}

	var scopeStrings []string
	if err := json.Unmarshal([]byte(scopesRaw), &scopeStrings); err != nil {
		return apitoken.APIToken{}, fmt.Errorf("unmarshal scopes: %w", err)
	}
	scopes := make(apitoken.Scopes, 0, len(scopeStrings))
	for _, s := range scopeStrings {
		sc, err := apitoken.NewScope(s)
		if err != nil {
			return apitoken.APIToken{}, fmt.Errorf("hydrate scope: %w", err)
		}
		scopes = append(scopes, sc)
	}

	return apitoken.ReconstituteAPIToken(
		tokenID, tenantID, userID, name, hash, prefix, scopes,
		createdAt.Time, expiresAt.Time, lastUsedAt.Time, isRevoked,
	)
}
