package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	prget "github.com/usegavel/gavel/core/application/pleading/get"
	prlist "github.com/usegavel/gavel/core/application/pleading/list"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

type PleadingFinder struct {
	db *database.DB
}

func NewPleadingFinder(db *database.DB) *PleadingFinder {
	return &PleadingFinder{db: db}
}

func (q *PleadingFinder) ListByProject(ctx context.Context, tenantID, projectID, status, gavelspace string, limit, offset int) ([]prlist.PleadingSummary, int, error) {
	where, args := buildPRWhere(tenantID, projectID, status, gavelspace)
	join := ""
	if gavelspace != "" {
		join = " INNER JOIN gavelspace_projects gp ON gp.project_id = pr.project_id"
	}

	var total int
	countSQL := "SELECT COUNT(*) FROM pleadings pr" + join + where
	if err := q.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count pull requests: %w", err)
	}

	dataSQL := `
		SELECT pr.id, pr.project_id, pr.number, pr.title, pr.petitioner,
		       pr.source_branch, pr.target_branch, pr.commit_sha, pr.status,
		       pr.created_at, pr.updated_at,
		       c.verdict_outcome
		FROM pleadings pr` + join + `
		LEFT JOIN (
			SELECT commit_sha, verdict_outcome,
			       ROW_NUMBER() OVER (PARTITION BY commit_sha ORDER BY created_at DESC) AS rn
			FROM casefiles
		) c ON c.commit_sha = pr.commit_sha AND c.rn = 1` +
		where + ` ORDER BY pr.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := q.db.QueryContext(ctx, dataSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list pull requests: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []prlist.PleadingSummary
	for rows.Next() {
		s, err := scanPRSummary(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, rows.Err()
}

func (q *PleadingFinder) GetByID(ctx context.Context, tenantID, pleadingID string) (*prget.PleadingDetail, error) {
	row := q.db.QueryRowContext(ctx, `
		SELECT pr.id, pr.project_id, pr.number, pr.title, pr.petitioner,
		       pr.source_branch, pr.target_branch, pr.commit_sha, pr.status,
		       pr.created_at, pr.updated_at,
		       c.id, c.verdict_outcome
		FROM pleadings pr
		LEFT JOIN (
			SELECT id, commit_sha, verdict_outcome,
			       ROW_NUMBER() OVER (PARTITION BY commit_sha ORDER BY created_at DESC) AS rn
			FROM casefiles
		) c ON c.commit_sha = pr.commit_sha AND c.rn = 1
		WHERE pr.id = ? AND pr.tenant_id = ?
	`, pleadingID, tenantID)

	var detail prget.PleadingDetail
	var createdAtStr, updatedAtStr string
	var caseFileID, verdictOutcome sql.NullString

	err := row.Scan(&detail.ID, &detail.ProjectID, &detail.Number, &detail.Title, &detail.Petitioner,
		&detail.SourceBranch, &detail.TargetBranch, &detail.CommitSHA, &detail.Status,
		&createdAtStr, &updatedAtStr,
		&caseFileID, &verdictOutcome)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: %s", errPleadingNotFound, pleadingID)
		}
		return nil, fmt.Errorf("get pull request: %w", err)
	}

	detail.CreatedAt, err = database.ParseTime(createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("pull request %s created_at: %w", detail.ID, err)
	}
	detail.UpdatedAt, err = database.ParseTime(updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("pull request %s updated_at: %w", detail.ID, err)
	}

	if verdictOutcome.Valid && verdictOutcome.String != "" {
		gateResult := &prget.GateResult{
			Passed: verdictOutcome.String == "pass",
		}
		if caseFileID.Valid {
			conditions, err := q.loadConditions(ctx, caseFileID.String)
			if err != nil {
				return nil, err
			}
			gateResult.Conditions = conditions
		}
		detail.GateResult = gateResult
	}

	return &detail, nil
}

func (q *PleadingFinder) loadConditions(ctx context.Context, caseFileID string) ([]prget.GateCondition, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT subtype, passed, detail
		FROM rulings WHERE casefile_id = ? ORDER BY sort_order
	`, caseFileID)
	if err != nil {
		return nil, fmt.Errorf("load rulings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var conditions []prget.GateCondition
	for rows.Next() {
		var subtype, detail string
		var passedInt int
		if err := rows.Scan(&subtype, &passedInt, &detail); err != nil {
			return nil, fmt.Errorf("scan ruling: %w", err)
		}
		conditions = append(conditions, rulingToCondition(subtype, passedInt != 0, detail))
	}
	return conditions, rows.Err()
}

var detailPatternPercent = regexp.MustCompile(`^([\d.]+%) .+ \(min ([\d.]+%)\)$`)
var detailPatternCount = regexp.MustCompile(`^(\d+) .+ \((?:max|threshold:?) (\d+)\)$`)

func rulingToCondition(subtype string, passed bool, detail string) prget.GateCondition {
	label := formatSubtypeLabel(subtype)
	operator := ">="
	value := detail
	threshold := ""

	if m := detailPatternPercent.FindStringSubmatch(detail); m != nil {
		value = m[1]
		threshold = m[2]
		operator = ">="
	} else if m := detailPatternCount.FindStringSubmatch(detail); m != nil {
		value = m[1]
		threshold = m[2]
		operator = "<="
	}

	return prget.GateCondition{
		Label:     label,
		Operator:  operator,
		Value:     value,
		Threshold: threshold,
		Passed:    passed,
	}
}

func formatSubtypeLabel(subtype string) string {
	s := strings.ReplaceAll(subtype, "_", " ")
	if len(s) > 0 {
		s = strings.ToUpper(s[:1]) + s[1:]
	}
	return s
}

func buildPRWhere(tenantID, projectID, status, gavelspace string) (string, []any) {
	conditions := []string{"pr.tenant_id = ?"}
	args := []any{tenantID}

	if projectID != "" {
		conditions = append(conditions, "pr.project_id = ?")
		args = append(args, projectID)
	}
	if status != "" {
		conditions = append(conditions, "pr.status = ?")
		args = append(args, status)
	}
	if gavelspace != "" {
		conditions = append(conditions, "gp.gavelspace_name = ?")
		args = append(args, gavelspace)
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

func scanPRSummary(rows *sql.Rows) (prlist.PleadingSummary, error) {
	var summary prlist.PleadingSummary
	var createdAtStr, updatedAtStr string
	var verdictOutcome sql.NullString

	err := rows.Scan(&summary.ID, &summary.ProjectID, &summary.Number, &summary.Title, &summary.Petitioner,
		&summary.SourceBranch, &summary.TargetBranch, &summary.CommitSHA, &summary.Status,
		&createdAtStr, &updatedAtStr,
		&verdictOutcome)
	if err != nil {
		return prlist.PleadingSummary{}, err
	}

	summary.CreatedAt, err = database.ParseTime(createdAtStr)
	if err != nil {
		return prlist.PleadingSummary{}, fmt.Errorf("pull request %s created_at: %w", summary.ID, err)
	}
	summary.UpdatedAt, err = database.ParseTime(updatedAtStr)
	if err != nil {
		return prlist.PleadingSummary{}, fmt.Errorf("pull request %s updated_at: %w", summary.ID, err)
	}

	if verdictOutcome.Valid && verdictOutcome.String != "" {
		summary.GateResult = &prlist.GateResult{
			Passed: verdictOutcome.String == "pass",
		}
	}

	return summary, nil
}
