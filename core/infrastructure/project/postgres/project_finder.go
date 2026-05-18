package postgres

import (
	"context"
	"database/sql"
	"fmt"

	projectget "github.com/usegavel/gavel/core/application/project/get"
	projectgetbykey "github.com/usegavel/gavel/core/application/project/getbykey"
	projectlist "github.com/usegavel/gavel/core/application/project/list"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

type ProjectFinder struct {
	db *database.DB
}

func NewProjectFinder(db *database.DB) *ProjectFinder {
	return &ProjectFinder{db: db}
}

func (q *ProjectFinder) List(ctx context.Context, limit, offset int) ([]projectlist.ProjectSummary, int, error) {
	var total int
	if err := q.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM projects").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}

	rows, err := q.db.QueryContext(ctx, `
		SELECT p.id, p.key, p.name, p.default_branch, p.created_at,
		       (SELECT verdict_outcome FROM casefiles
		        WHERE project_id = p.id ORDER BY started_at DESC LIMIT 1) as latest_verdict,
		       (SELECT total_findings FROM casefiles
		        WHERE project_id = p.id ORDER BY started_at DESC LIMIT 1) as total_findings
		FROM projects p
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var items []projectlist.ProjectSummary
	for rows.Next() {
		var summary projectlist.ProjectSummary
		var createdAtStr string
		var latestVerdict sql.NullString
		var totalFindings sql.NullInt64
		if err := rows.Scan(&summary.ID, &summary.Key, &summary.Name, &summary.DefaultBranch, &createdAtStr,
			&latestVerdict, &totalFindings); err != nil {
			return nil, 0, err
		}
		summary.CreatedAt, err = database.ParseTime(createdAtStr)
		if err != nil {
			return nil, 0, fmt.Errorf("project %s created_at: %w", summary.ID, err)
		}
		summary.LatestVerdict = latestVerdict.String
		summary.TotalFindings = int(totalFindings.Int64)
		items = append(items, summary)
	}
	return items, total, rows.Err()
}

func (q *ProjectFinder) GetByID(ctx context.Context, id string) (*projectget.ProjectDetail, error) {
	raw, err := q.getProject(ctx, "SELECT p.id, p.key, p.name, p.target_pattern, p.default_branch, p.created_at FROM projects p WHERE p.id = ?", id)
	if err != nil {
		return nil, err
	}
	return &projectget.ProjectDetail{
		ID:               raw.id,
		Key:              raw.key,
		Name:             raw.name,
		TargetPattern:    raw.targetPattern,
		DefaultBranch:    raw.defaultBranch,
		LatestVerdict:    raw.latestVerdict,
		TotalFindings:    raw.totalFindings,
		CreatedAt:        raw.createdAt,
		Languages:        raw.languages,
		QualityGateRules: toGetQGRules(raw.qgRules),
		SeverityCounts:   raw.severityCounts,
	}, nil
}

func (q *ProjectFinder) GetByKey(ctx context.Context, key string) (*projectgetbykey.ProjectDetail, error) {
	raw, err := q.getProject(ctx, "SELECT p.id, p.key, p.name, p.target_pattern, p.default_branch, p.created_at FROM projects p WHERE p.key = ?", key)
	if err != nil {
		return nil, err
	}
	return &projectgetbykey.ProjectDetail{
		ID:               raw.id,
		Key:              raw.key,
		Name:             raw.name,
		TargetPattern:    raw.targetPattern,
		DefaultBranch:    raw.defaultBranch,
		LatestVerdict:    raw.latestVerdict,
		TotalFindings:    raw.totalFindings,
		CreatedAt:        raw.createdAt,
		Languages:        raw.languages,
		QualityGateRules: toGetByKeyQGRules(raw.qgRules),
		SeverityCounts:   raw.severityCounts,
	}, nil
}

func (q *ProjectFinder) getProject(ctx context.Context, query, param string) (*projectDetailRow, error) {
	var detail projectDetailRow
	var createdAtStr string
	err := q.db.QueryRowContext(ctx, query, param).Scan(
		&detail.id, &detail.key, &detail.name, &detail.targetPattern, &detail.defaultBranch, &createdAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: %s", errProjectNotFound, param)
		}
		return nil, err
	}
	detail.createdAt, err = database.ParseTime(createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("project %s created_at: %w", detail.id, err)
	}

	languages, err := q.loadLanguageStrings(ctx, detail.id)
	if err != nil {
		return nil, err
	}
	detail.languages = languages

	qgRules, err := q.loadQGRuleRows(ctx, detail.id)
	if err != nil {
		return nil, err
	}
	detail.qgRules = qgRules

	severityCounts, err := q.loadLatestSeverityCounts(ctx, detail.id)
	if err != nil {
		return nil, err
	}
	detail.severityCounts = severityCounts

	latestVerdict, totalFindings, err := q.loadLatestCaseFileSummary(ctx, detail.id)
	if err != nil {
		return nil, err
	}
	detail.latestVerdict = latestVerdict
	detail.totalFindings = totalFindings

	return &detail, nil
}

func (q *ProjectFinder) loadLanguageStrings(ctx context.Context, projectID string) ([]string, error) {
	rows, err := q.db.QueryContext(ctx,
		"SELECT language FROM project_languages WHERE project_id = ?", projectID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	langs := make([]string, 0)
	for rows.Next() {
		var l string
		if err := rows.Scan(&l); err != nil {
			return nil, err
		}
		langs = append(langs, l)
	}
	return langs, rows.Err()
}

func (q *ProjectFinder) loadQGRuleRows(ctx context.Context, projectID string) ([]qgRuleRow, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT subtype, strategy_type
		FROM project_quality_gate_rules WHERE project_id = ? ORDER BY sort_order
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	rules := make([]qgRuleRow, 0)
	for rows.Next() {
		var rv qgRuleRow
		if err := rows.Scan(&rv.subtype, &rv.strategyType); err != nil {
			return nil, err
		}
		rules = append(rules, rv)
	}
	return rules, rows.Err()
}

func (q *ProjectFinder) loadLatestSeverityCounts(ctx context.Context, projectID string) (map[string]int, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT severity, COUNT(*) FROM findings
		WHERE casefile_id = (
		    SELECT id FROM casefiles WHERE project_id = ? ORDER BY started_at DESC LIMIT 1
		)
		GROUP BY severity
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	counts := make(map[string]int)
	for rows.Next() {
		var sev string
		var cnt int
		if err := rows.Scan(&sev, &cnt); err != nil {
			return nil, err
		}
		counts[sev] = cnt
	}
	return counts, rows.Err()
}

func (q *ProjectFinder) loadLatestCaseFileSummary(ctx context.Context, projectID string) (string, int, error) {
	var verdict sql.NullString
	var total sql.NullInt64
	err := q.db.QueryRowContext(ctx, `
		SELECT verdict_outcome, total_findings FROM casefiles
		WHERE project_id = ? ORDER BY started_at DESC LIMIT 1
	`, projectID).Scan(&verdict, &total)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", 0, nil
		}
		return "", 0, err
	}
	return verdict.String, int(total.Int64), nil
}

func toGetQGRules(rows []qgRuleRow) []projectget.QualityGateRuleView {
	rules := make([]projectget.QualityGateRuleView, len(rows))
	for i, r := range rows {
		rules[i] = projectget.QualityGateRuleView{
			Subtype:      r.subtype,
			StrategyType: r.strategyType,
		}
	}
	return rules
}

func toGetByKeyQGRules(rows []qgRuleRow) []projectgetbykey.QualityGateRuleView {
	rules := make([]projectgetbykey.QualityGateRuleView, len(rows))
	for i, r := range rows {
		rules[i] = projectgetbykey.QualityGateRuleView{
			Subtype:      r.subtype,
			StrategyType: r.strategyType,
		}
	}
	return rules
}
