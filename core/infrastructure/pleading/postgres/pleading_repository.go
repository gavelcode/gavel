package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/pleading/model"
	pleadingservice "github.com/usegavel/gavel/core/domain/pleading/service"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/shared/failure"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

var _ pleadingservice.PleadingRepository = (*Repository)(nil)

var errPleadingNotFound = failure.New("pleading not found", failure.NotFound)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(ctx context.Context, pleading model.Pleading) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE pleadings SET
			project_id = ?,
			number = ?,
			title = ?,
			petitioner = ?,
			status = ?,
			source_branch = ?,
			target_branch = ?,
			commit_sha = ?,
			updated_at = ?
		WHERE id = ?
	`,
		pleading.ProjectID().UUID(), pleading.Number(), pleading.Title(), pleading.Petitioner(),
		pleading.Status().String(), pleading.SourceBranch(), pleading.TargetBranch(), pleading.CommitSHA(),
		database.Now(), pleading.ID().UUID(),
	)
	if err != nil {
		return fmt.Errorf("update pleading: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows > 0 {
		return nil
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO pleadings (id, project_id, tenant_id, number, title, petitioner, status, source_branch, target_branch, commit_sha)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(project_id, number) DO UPDATE SET
			title = EXCLUDED.title,
			petitioner = EXCLUDED.petitioner,
			source_branch = EXCLUDED.source_branch,
			target_branch = EXCLUDED.target_branch,
			commit_sha = EXCLUDED.commit_sha,
			updated_at = ?
	`,
		pleading.ID().UUID(), pleading.ProjectID().UUID(), pleading.TenantID().UUID(), pleading.Number(), pleading.Title(),
		pleading.Petitioner(), pleading.Status().String(), pleading.SourceBranch(), pleading.TargetBranch(), pleading.CommitSHA(),
		database.Now(),
	)
	if err != nil {
		return fmt.Errorf("insert pleading: %w", err)
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, tenantID tenant.TenantID, pleadingID model.PleadingID) (model.Pleading, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, project_id, number, title, petitioner, status, source_branch, target_branch, commit_sha
		FROM pleadings WHERE id = ? AND tenant_id = ?
	`, pleadingID.UUID(), tenantID.UUID())

	var (
		rawID, rawProjectID                                                 uuid.UUID
		title, petitioner, statusStr, sourceBranch, targetBranch, commitSHA string
		number                                                              int
	)
	if err := row.Scan(&rawID, &rawProjectID, &number, &title, &petitioner, &statusStr, &sourceBranch, &targetBranch, &commitSHA); err != nil {
		if err == sql.ErrNoRows {
			return model.Pleading{}, fmt.Errorf("%w: %s", errPleadingNotFound, pleadingID.String())
		}
		return model.Pleading{}, fmt.Errorf("scan pleading: %w", err)
	}

	pleadingID = model.NewPleadingID(rawID)
	projectID := projectmodel.NewProjectID(rawProjectID)
	status, err := model.NewStatus(statusStr)
	if err != nil {
		return model.Pleading{}, fmt.Errorf("reconstitute pleading status: %w", err)
	}

	return model.ReconstitutePleading(pleadingID, tenantID, projectID, number, title, petitioner, sourceBranch, targetBranch, commitSHA, status)
}
