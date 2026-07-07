package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	tenantmodel "github.com/usegavel/gavel/core/domain/iam/model/tenant"
	usermodel "github.com/usegavel/gavel/core/domain/iam/model/user"
	"github.com/usegavel/gavel/core/domain/iam/service"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

type UserRepo struct {
	db database.Querier
}

var _ service.UserRepository = (*UserRepo)(nil)

func NewUserRepo(db database.Querier) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Save(ctx context.Context, user usermodel.User) error {
	var lastLogin any
	if !user.LastLoginAt().IsZero() {
		lastLogin = user.LastLoginAt()
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO iam_users (
			id, tenant_id, email, display_name, role, password_hash,
			must_change_password, is_active, created_at, last_login_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			role = EXCLUDED.role,
			password_hash = EXCLUDED.password_hash,
			must_change_password = EXCLUDED.must_change_password,
			is_active = EXCLUDED.is_active,
			last_login_at = EXCLUDED.last_login_at
	`,
		user.ID().UUID(), user.TenantID().UUID(), user.Email().String(),
		user.DisplayName(), user.Role().String(), user.PasswordHash().String(),
		user.MustChangePassword(), user.IsActive(), user.CreatedAt(), lastLogin,
	)
	if err != nil {
		if isUniqueViolation(err, "iam_users_tenant_id_email_key") {
			return fmt.Errorf("%w: %s", usermodel.ErrEmailAlreadyInUse, user.Email())
		}
		return fmt.Errorf("save user: %w", err)
	}
	return nil
}

func (r *UserRepo) ByID(ctx context.Context, id usermodel.UserID) (usermodel.User, error) {
	return r.scanOne(ctx, `WHERE id = ?`, id.UUID())
}

func (r *UserRepo) ByEmail(ctx context.Context, tenantID tenantmodel.TenantID, email usermodel.Email) (usermodel.User, error) {
	return r.scanOne(ctx, `WHERE tenant_id = ? AND email = ?`, tenantID.UUID(), email.String())
}

func (r *UserRepo) CountByTenant(ctx context.Context, tenantID tenantmodel.TenantID) (int, error) {
	row := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM iam_users WHERE tenant_id = ?`, tenantID.UUID())
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return n, nil
}

func (r *UserRepo) scanOne(ctx context.Context, where string, args ...any) (usermodel.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, email, display_name, role, password_hash,
		       must_change_password, is_active, created_at, last_login_at
		FROM iam_users
	`+where, args...)

	var idVal, tenantIDVal uuid.UUID
	var emailRaw, displayName, roleRaw, hashRaw string
	var mustChange, isActive bool
	var createdAt sql.NullTime
	var lastLoginAt sql.NullTime
	if err := row.Scan(&idVal, &tenantIDVal, &emailRaw, &displayName, &roleRaw, &hashRaw, &mustChange, &isActive, &createdAt, &lastLoginAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return usermodel.User{}, fmt.Errorf("%w", usermodel.ErrUserNotFound)
		}
		return usermodel.User{}, fmt.Errorf("scan user: %w", err)
	}

	userID := usermodel.NewUserID(idVal)
	tenantID := tenantmodel.NewTenantID(tenantIDVal)
	email, err := usermodel.NewEmail(emailRaw)
	if err != nil {
		return usermodel.User{}, fmt.Errorf("hydrate email: %w", err)
	}
	role, err := usermodel.NewRole(roleRaw)
	if err != nil {
		return usermodel.User{}, fmt.Errorf("hydrate role: %w", err)
	}
	hash, err := usermodel.NewPasswordHash(hashRaw)
	if err != nil {
		return usermodel.User{}, fmt.Errorf("hydrate password hash: %w", err)
	}

	var lastLogin = lastLoginAt.Time
	if !lastLoginAt.Valid {
		var zero = lastLogin
		lastLogin = zero
	}

	return usermodel.ReconstituteUser(userID, tenantID, email, displayName, role, hash, mustChange, isActive, createdAt.Time, lastLogin)
}
