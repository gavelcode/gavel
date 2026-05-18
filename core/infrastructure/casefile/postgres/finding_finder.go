package postgres

import (
	"context"
	"fmt"
	"strings"

	findinglist "github.com/usegavel/gavel/core/application/casefile/listfindings"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

type FindingFinder struct {
	db *database.DB
}

func NewFindingFinder(db *database.DB) *FindingFinder {
	return &FindingFinder{db: db}
}

func (q *FindingFinder) List(ctx context.Context, filters findinglist.Filters, limit, offset int) ([]findinglist.FindingView, int, error) {
	where, args := buildFindingWhere(filters)
	join := ""
	if filters.Gavelspace != "" {
		join = " INNER JOIN gavelspace_projects gp ON gp.project_id = f.project_id"
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM findings f" + join + where
	if err := q.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count findings: %w", err)
	}

	dataQuery := `
		SELECT f.tool, f.rule_id, f.severity, f.file_path, f.line,
		       f.message, f.fingerprint, f.status,
		       COALESCE(e.source, '') as source,
		       COALESCE(cf.commit_sha, '') as commit_sha,
		       COALESCE(p.key, '') as project_key,
		       COALESCE(f.casefile_id::text, '') as casefile_id
		FROM findings f` + join + `
		LEFT JOIN evidences e ON f.evidence_id = e.id
		LEFT JOIN casefiles cf ON cf.id = f.casefile_id
		LEFT JOIN projects p ON p.id = cf.project_id` + where + `
		ORDER BY f.created_at DESC
		LIMIT ? OFFSET ?`
	dataArgs := append(args, limit, offset)

	rows, err := q.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var items []findinglist.FindingView
	for rows.Next() {
		var findingView findinglist.FindingView
		if err := rows.Scan(&findingView.Tool, &findingView.RuleID, &findingView.Severity, &findingView.FilePath,
			&findingView.Line, &findingView.Message, &findingView.FingerprintID, &findingView.Status, &findingView.Source,
			&findingView.CommitSHA, &findingView.ProjectKey, &findingView.CaseFileID); err != nil {
			return nil, 0, err
		}
		items = append(items, findingView)
	}
	return items, total, rows.Err()
}

func (q *FindingFinder) ListByFile(ctx context.Context, caseFileID, filePath string) ([]findinglist.FindingView, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT f.tool, f.rule_id, f.severity, f.file_path, f.line,
		       f.message, f.fingerprint, f.status,
		       COALESCE(e.source, '') as source,
		       COALESCE(cf.commit_sha, '') as commit_sha,
		       COALESCE(p.key, '') as project_key,
		       COALESCE(f.casefile_id::text, '') as casefile_id
		FROM findings f
		LEFT JOIN evidences e ON f.evidence_id = e.id
		LEFT JOIN casefiles cf ON cf.id = f.casefile_id
		LEFT JOIN projects p ON p.id = cf.project_id
		WHERE f.casefile_id = ? AND f.file_path = ?
		ORDER BY f.line ASC
	`, caseFileID, filePath)
	if err != nil {
		return nil, fmt.Errorf("list findings by file: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []findinglist.FindingView
	for rows.Next() {
		var findingView findinglist.FindingView
		if err := rows.Scan(&findingView.Tool, &findingView.RuleID, &findingView.Severity, &findingView.FilePath,
			&findingView.Line, &findingView.Message, &findingView.FingerprintID, &findingView.Status, &findingView.Source,
			&findingView.CommitSHA, &findingView.ProjectKey, &findingView.CaseFileID); err != nil {
			return nil, err
		}
		items = append(items, findingView)
	}
	return items, rows.Err()
}

func buildFindingWhere(filters findinglist.Filters) (string, []any) {
	var conditions []string
	var args []any

	if filters.ProjectID != "" {
		conditions = append(conditions, "f.project_id = ?")
		args = append(args, filters.ProjectID)
	}
	if filters.CaseFileID != "" {
		conditions = append(conditions, "f.casefile_id = ?")
		args = append(args, filters.CaseFileID)
	}
	if filters.Tool != "" {
		conditions = append(conditions, "f.tool = ?")
		args = append(args, filters.Tool)
	}
	if filters.Severity != "" {
		conditions = append(conditions, "f.severity = ?")
		args = append(args, filters.Severity)
	}
	if filters.Status != "" {
		conditions = append(conditions, "f.status = ?")
		args = append(args, filters.Status)
	}
	if filters.FilePath != "" {
		conditions = append(conditions, "f.file_path LIKE ? ESCAPE '\\'")
		args = append(args, database.EscapeLike(filters.FilePath)+"%")
	}
	if filters.Gavelspace != "" {
		conditions = append(conditions, "gp.gavelspace_name = ?")
		args = append(args, filters.Gavelspace)
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}
