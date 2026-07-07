package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	tenantmodel "github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/iam/service"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

type TenantRepo struct {
	db database.Querier
}

var _ service.TenantRepository = (*TenantRepo)(nil)

func NewTenantRepo(db database.Querier) *TenantRepo {
	return &TenantRepo{db: db}
}

func (r *TenantRepo) Save(ctx context.Context, tenant tenantmodel.Tenant) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO iam_tenants (id, slug, display_name, status, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
		    slug = EXCLUDED.slug,
		    display_name = EXCLUDED.display_name,
		    status = EXCLUDED.status
	`,
		tenant.ID().UUID(),
		tenant.Slug().String(),
		tenant.DisplayName(),
		tenant.Status().String(),
		tenant.CreatedAt(),
	)
	if err != nil {
		if isUniqueViolation(err, "iam_tenants_slug_key") {
			return fmt.Errorf("%w: %s", tenantmodel.ErrSlugTaken, tenant.Slug())
		}
		return fmt.Errorf("save tenant: %w", err)
	}
	return nil
}

func (r *TenantRepo) ByID(ctx context.Context, id tenantmodel.TenantID) (tenantmodel.Tenant, error) {
	return r.scanOne(ctx, `WHERE id = ?`, id.UUID())
}

func (r *TenantRepo) BySlug(ctx context.Context, slug tenantmodel.Slug) (tenantmodel.Tenant, error) {
	return r.scanOne(ctx, `WHERE slug = ?`, slug.String())
}

func (r *TenantRepo) scanOne(ctx context.Context, where string, args ...any) (tenantmodel.Tenant, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, slug, display_name, status, created_at
		FROM iam_tenants
	`+where, args...)

	var idVal uuid.UUID
	var slugRaw, displayName, statusRaw string
	var createdAt sql.NullTime
	if err := row.Scan(&idVal, &slugRaw, &displayName, &statusRaw, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tenantmodel.Tenant{}, fmt.Errorf("%w", tenantmodel.ErrTenantNotFound)
		}
		return tenantmodel.Tenant{}, fmt.Errorf("scan tenant: %w", err)
	}

	tenantID := tenantmodel.NewTenantID(idVal)
	slug, err := tenantmodel.NewSlug(slugRaw)
	if err != nil {
		return tenantmodel.Tenant{}, fmt.Errorf("hydrate tenant slug: %w", err)
	}
	status, err := tenantmodel.NewStatus(statusRaw)
	if err != nil {
		return tenantmodel.Tenant{}, fmt.Errorf("hydrate tenant status: %w", err)
	}
	return tenantmodel.ReconstituteTenant(tenantID, slug, displayName, status, createdAt.Time)
}
