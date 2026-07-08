package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	gsmodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/shared/failure"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

var errGavelspaceNotFound = failure.New("gavelspace not found", failure.NotFound)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(ctx context.Context, gavelspace gsmodel.Gavelspace) error {
	transaction, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	_, err = transaction.ExecContext(ctx, `
		INSERT INTO gavelspaces (name, tenant_id)
		VALUES (?, ?)
		ON CONFLICT(tenant_id, name) DO UPDATE SET
		    updated_at = ?
	`, gavelspace.ID().String(), gavelspace.TenantID().UUID(), database.Now())
	if err != nil {
		return fmt.Errorf("upsert gavelspace: %w", err)
	}

	if err := r.replaceProjectRefs(ctx, transaction, gavelspace); err != nil {
		return fmt.Errorf("replace project refs: %w", err)
	}

	return transaction.Commit()
}

func (r *Repository) replaceProjectRefs(ctx context.Context, transaction *database.Tx, gavelspace gsmodel.Gavelspace) error {
	_, err := transaction.ExecContext(ctx,
		"DELETE FROM gavelspace_projects WHERE gavelspace_name = ? AND tenant_id = ?",
		gavelspace.ID().String(), gavelspace.TenantID().UUID())
	if err != nil {
		return err
	}

	for _, ref := range gavelspace.Projects() {
		_, err = transaction.ExecContext(ctx, `
			INSERT INTO gavelspace_projects (gavelspace_name, tenant_id, project_id, target_pattern)
			VALUES (?, ?, ?, ?)
		`, gavelspace.ID().String(), gavelspace.TenantID().UUID(), ref.ID().UUID(), ref.TargetPattern())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) FindByName(ctx context.Context, tenantID tenant.TenantID, name gsmodel.GavelspaceID) (gsmodel.Gavelspace, error) {
	var nameStr string
	var tenantIDVal uuid.UUID
	err := r.db.QueryRowContext(ctx,
		"SELECT name, tenant_id FROM gavelspaces WHERE name = ? AND tenant_id = ?",
		name.String(), tenantID.UUID()).Scan(&nameStr, &tenantIDVal)
	if err != nil {
		if err == sql.ErrNoRows {
			return gsmodel.Gavelspace{}, fmt.Errorf("%w: %s", errGavelspaceNotFound, name)
		}
		return gsmodel.Gavelspace{}, fmt.Errorf("query gavelspace: %w", err)
	}

	refs, err := r.loadProjectRefs(ctx, tenantID, nameStr)
	if err != nil {
		return gsmodel.Gavelspace{}, fmt.Errorf("load project refs: %w", err)
	}

	return gsmodel.ReconstituteGavelspace(name, tenant.NewTenantID(tenantIDVal), refs)
}

func (r *Repository) loadProjectRefs(ctx context.Context, tenantID tenant.TenantID, gavelspaceName string) ([]gsmodel.ProjectRef, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT project_id, target_pattern
		FROM gavelspace_projects WHERE gavelspace_name = ? AND tenant_id = ?
	`, gavelspaceName, tenantID.UUID())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var refs []gsmodel.ProjectRef
	for rows.Next() {
		var projectIDVal uuid.UUID
		var targetPattern string
		if err := rows.Scan(&projectIDVal, &targetPattern); err != nil {
			return nil, err
		}
		projectID := projectmodel.NewProjectID(projectIDVal)
		ref, err := gsmodel.NewProjectRef(projectID, targetPattern)
		if err != nil {
			return nil, fmt.Errorf("reconstitute project ref: %w", err)
		}
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}
