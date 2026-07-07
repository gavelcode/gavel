package postgres

import (
	"context"
	"database/sql"
	"fmt"

	gsget "github.com/usegavel/gavel/core/application/gavelspace/get"
	gslist "github.com/usegavel/gavel/core/application/gavelspace/list"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

type GavelspaceFinder struct {
	db *database.DB
}

func NewGavelspaceFinder(db *database.DB) *GavelspaceFinder {
	return &GavelspaceFinder{db: db}
}

func (q *GavelspaceFinder) List(ctx context.Context, tenantID tenant.TenantID, limit, offset int) ([]gslist.GavelspaceSummary, int, error) {
	var total int
	if err := q.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM gavelspaces WHERE tenant_id = ?", tenantID.UUID()).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count gavelspaces: %w", err)
	}

	rows, err := q.db.QueryContext(ctx, `
		SELECT g.name,
		       (SELECT COUNT(*) FROM gavelspace_projects WHERE gavelspace_name = g.name) AS project_count,
		       g.created_at
		FROM gavelspaces g
		WHERE g.tenant_id = ?
		ORDER BY g.name
		LIMIT ? OFFSET ?
	`, tenantID.UUID(), limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list gavelspaces: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []gslist.GavelspaceSummary
	for rows.Next() {
		var summary gslist.GavelspaceSummary
		var createdAtStr string
		if err := rows.Scan(&summary.Name, &summary.ProjectCount, &createdAtStr); err != nil {
			return nil, 0, err
		}
		summary.CreatedAt, err = database.ParseTime(createdAtStr)
		if err != nil {
			return nil, 0, fmt.Errorf("gavelspace %s created_at: %w", summary.Name, err)
		}
		items = append(items, summary)
	}
	return items, total, rows.Err()
}

func (q *GavelspaceFinder) GetByName(ctx context.Context, tenantID tenant.TenantID, name string) (*gsget.GavelspaceDetail, error) {
	var detail gsget.GavelspaceDetail
	var createdAtStr string
	err := q.db.QueryRowContext(ctx,
		"SELECT name, created_at FROM gavelspaces WHERE name = ? AND tenant_id = ?",
		name, tenantID.UUID()).Scan(&detail.Name, &createdAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: %s", errGavelspaceNotFound, name)
		}
		return nil, fmt.Errorf("get gavelspace: %w", err)
	}
	detail.CreatedAt, err = database.ParseTime(createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("gavelspace %s created_at: %w", detail.Name, err)
	}

	projects, err := q.loadProjectRefViews(ctx, name)
	if err != nil {
		return nil, err
	}
	detail.Projects = projects

	return &detail, nil
}

func (q *GavelspaceFinder) FindGavelspaceNameByProjectID(ctx context.Context, projectID string) (string, bool, error) {
	var name string
	err := q.db.QueryRowContext(ctx,
		"SELECT gavelspace_name FROM gavelspace_projects WHERE project_id = ? LIMIT 1",
		projectID).Scan(&name)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("find gavelspace by project: %w", err)
	}
	return name, true, nil
}

func (q *GavelspaceFinder) loadProjectRefViews(ctx context.Context, gavelspaceName string) ([]gsget.ProjectRefView, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT gp.project_id, p.key, p.name,
		       (SELECT verdict_outcome FROM casefiles
		        WHERE project_id = gp.project_id
		        ORDER BY started_at DESC LIMIT 1) AS latest_verdict
		FROM gavelspace_projects gp
		JOIN projects p ON p.id = gp.project_id
		WHERE gp.gavelspace_name = ?
		ORDER BY p.name
	`, gavelspaceName)
	if err != nil {
		return nil, fmt.Errorf("load project refs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var views []gsget.ProjectRefView
	for rows.Next() {
		var view gsget.ProjectRefView
		var latestVerdict sql.NullString
		if err := rows.Scan(&view.ID, &view.Key, &view.Name, &latestVerdict); err != nil {
			return nil, err
		}
		view.LatestVerdict = latestVerdict.String
		views = append(views, view)
	}
	return views, rows.Err()
}
