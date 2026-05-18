package postgres

import (
	"context"
	"database/sql"
	"fmt"

	casefileget "github.com/usegavel/gavel/core/application/casefile/get"
	casefilelist "github.com/usegavel/gavel/core/application/casefile/list"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

type CaseFileFinder struct {
	db *database.DB
}

func NewCaseFileFinder(db *database.DB) *CaseFileFinder {
	return &CaseFileFinder{db: db}
}

func (q *CaseFileFinder) ListByProject(ctx context.Context, projectID, gavelspace string, limit, offset int) ([]casefilelist.CaseFileSummary, int, error) {
	join, where, args := buildCaseFileFilter(projectID, gavelspace)

	var total int
	if err := q.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM casefiles c"+join+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count casefiles: %w", err)
	}

	rows, err := q.db.QueryContext(ctx, `
		SELECT c.id, c.project_id, c.commit_sha, c.branch, c.started_at,
		       c.verdict_outcome, c.total_findings, c.new_findings, c.existing_findings,
		       c.resolved_findings,
		       CASE WHEN cd.total_lines > 0
		            THEN (cd.covered_lines * 100.0) / cd.total_lines
		            ELSE NULL END,
		       c.created_at
		FROM casefiles c`+join+`
		LEFT JOIN evidences e ON e.casefile_id = c.id AND e.subtype = 'coverage'
		LEFT JOIN coverage_data cd ON cd.evidence_id = e.id`+where+`
		ORDER BY c.created_at DESC
		LIMIT ? OFFSET ?
	`, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var items []casefilelist.CaseFileSummary
	for rows.Next() {
		s, err := scanCaseFileSummary(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, rows.Err()
}

func buildCaseFileFilter(projectID, gavelspace string) (join, where string, args []any) {
	var conditions []string
	if gavelspace != "" {
		join = " INNER JOIN gavelspace_projects gp ON gp.project_id = c.project_id"
		conditions = append(conditions, "gp.gavelspace_name = ?")
		args = append(args, gavelspace)
	}
	if projectID != "" {
		conditions = append(conditions, "c.project_id = ?")
		args = append(args, projectID)
	}
	if len(conditions) == 0 {
		return "", "", nil
	}
	where = " WHERE " + conditions[0]
	for _, c := range conditions[1:] {
		where += " AND " + c
	}
	return join, where, args
}

func (q *CaseFileFinder) GetByID(ctx context.Context, caseFileID string) (*casefileget.CaseFileDetail, error) {
	var detail casefileget.CaseFileDetail
	var verdictOutcome sql.NullString
	var coveragePct sql.NullFloat64
	var startedAtStr, createdAtStr string

	err := q.db.QueryRowContext(ctx, `
		SELECT c.id, c.project_id, c.commit_sha, c.branch, c.started_at,
		       c.verdict_outcome, c.total_findings, c.new_findings, c.existing_findings,
		       c.resolved_findings,
		       CASE WHEN cd.total_lines > 0
		            THEN (cd.covered_lines * 100.0) / cd.total_lines
		            ELSE NULL END,
		       c.created_at
		FROM casefiles c
		LEFT JOIN evidences e ON e.casefile_id = c.id AND e.subtype = 'coverage'
		LEFT JOIN coverage_data cd ON cd.evidence_id = e.id
		WHERE c.id = ?
	`, caseFileID).Scan(&detail.ID, &detail.ProjectID, &detail.CommitSHA, &detail.Branch, &startedAtStr,
		&verdictOutcome, &detail.TotalFindings, &detail.NewFindings, &detail.ExistingFindings,
		&detail.ResolvedFindings, &coveragePct, &createdAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: %s", errCaseFileNotFound, caseFileID)
		}
		return nil, err
	}

	detail.VerdictOutcome = verdictOutcome.String
	if coveragePct.Valid {
		detail.CoveragePercent = &coveragePct.Float64
	}
	detail.StartedAt, err = database.ParseTime(startedAtStr)
	if err != nil {
		return nil, fmt.Errorf("case file %s started_at: %w", detail.ID, err)
	}
	detail.CreatedAt, err = database.ParseTime(createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("case file %s created_at: %w", detail.ID, err)
	}

	evidences, err := q.loadEvidenceSummaries(ctx, caseFileID)
	if err != nil {
		return nil, err
	}
	detail.Evidences = evidences

	rulings, err := q.loadRulingViews(ctx, caseFileID)
	if err != nil {
		return nil, err
	}
	detail.Rulings = rulings

	return &detail, nil
}

func (q *CaseFileFinder) loadEvidenceSummaries(ctx context.Context, caseFileID string) ([]casefileget.EvidenceSummary, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT id, subtype, source, collected_at
		FROM evidences WHERE casefile_id = ?
	`, caseFileID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []casefileget.EvidenceSummary
	for rows.Next() {
		var summary casefileget.EvidenceSummary
		var collectedAtStr string
		if err := rows.Scan(&summary.ID, &summary.Subtype, &summary.Source, &collectedAtStr); err != nil {
			return nil, err
		}
		summary.CollectedAt, err = database.ParseTime(collectedAtStr)
		if err != nil {
			return nil, fmt.Errorf("evidence %s collected_at: %w", summary.ID, err)
		}
		items = append(items, summary)
	}
	return items, rows.Err()
}

func (q *CaseFileFinder) loadRulingViews(ctx context.Context, caseFileID string) ([]casefileget.RulingView, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT subtype, passed, detail
		FROM rulings WHERE casefile_id = ? ORDER BY sort_order
	`, caseFileID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []casefileget.RulingView
	for rows.Next() {
		var ruling casefileget.RulingView
		var passedInt int
		if err := rows.Scan(&ruling.Subtype, &passedInt, &ruling.Detail); err != nil {
			return nil, err
		}
		ruling.Passed = passedInt != 0
		items = append(items, ruling)
	}
	return items, rows.Err()
}

func scanCaseFileSummary(rows *sql.Rows) (casefilelist.CaseFileSummary, error) {
	var summary casefilelist.CaseFileSummary
	var verdictOutcome sql.NullString
	var coveragePct sql.NullFloat64
	var startedAtStr, createdAtStr string

	err := rows.Scan(&summary.ID, &summary.ProjectID, &summary.CommitSHA, &summary.Branch, &startedAtStr,
		&verdictOutcome, &summary.TotalFindings, &summary.NewFindings, &summary.ExistingFindings,
		&summary.ResolvedFindings, &coveragePct, &createdAtStr)
	if err != nil {
		return casefilelist.CaseFileSummary{}, err
	}

	summary.VerdictOutcome = verdictOutcome.String
	if coveragePct.Valid {
		summary.CoveragePercent = &coveragePct.Float64
	}
	summary.StartedAt, err = database.ParseTime(startedAtStr)
	if err != nil {
		return casefilelist.CaseFileSummary{}, fmt.Errorf("case file %s started_at: %w", summary.ID, err)
	}
	summary.CreatedAt, err = database.ParseTime(createdAtStr)
	if err != nil {
		return casefilelist.CaseFileSummary{}, fmt.Errorf("case file %s created_at: %w", summary.ID, err)
	}

	return summary, nil
}
