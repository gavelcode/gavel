package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	gsmodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
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
		INSERT INTO gavelspaces (name)
		VALUES (?)
		ON CONFLICT(name) DO UPDATE SET
		    updated_at = ?
	`, gavelspace.ID().String(), database.Now())
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
		"DELETE FROM gavelspace_projects WHERE gavelspace_name = ?",
		gavelspace.ID().String())
	if err != nil {
		return err
	}

	for _, ref := range gavelspace.Projects() {
		_, err = transaction.ExecContext(ctx, `
			INSERT INTO gavelspace_projects (gavelspace_name, project_id, target_pattern)
			VALUES (?, ?, ?)
		`, gavelspace.ID().String(), ref.ID().UUID(), ref.TargetPattern())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) FindByName(ctx context.Context, name gsmodel.GavelspaceID) (gsmodel.Gavelspace, error) {
	var nameStr string
	err := r.db.QueryRowContext(ctx,
		"SELECT name FROM gavelspaces WHERE name = ?",
		name.String()).Scan(&nameStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return gsmodel.Gavelspace{}, fmt.Errorf("%w: %s", errGavelspaceNotFound, name)
		}
		return gsmodel.Gavelspace{}, fmt.Errorf("query gavelspace: %w", err)
	}

	refs, err := r.loadProjectRefs(ctx, nameStr)
	if err != nil {
		return gsmodel.Gavelspace{}, fmt.Errorf("load project refs: %w", err)
	}

	return gsmodel.ReconstituteGavelspace(name, refs)
}

func (r *Repository) loadProjectRefs(ctx context.Context, gavelspaceName string) ([]gsmodel.ProjectRef, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT project_id, target_pattern
		FROM gavelspace_projects WHERE gavelspace_name = ?
	`, gavelspaceName)
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
