package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/usegavel/gavel/core/application/casefile/finalize"
	"github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/architecture"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/toolexecution"
	"github.com/usegavel/gavel/core/domain/casefile/model/verdict"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/shared/failure"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

var errCaseFileNotFound = failure.New("case file not found", failure.NotFound)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(ctx context.Context, caseFile model.CaseFile) error {
	transaction, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	if err := r.upsertCaseFile(ctx, transaction, caseFile); err != nil {
		return fmt.Errorf("upsert casefile: %w", err)
	}

	for _, ev := range caseFile.Evidences() {
		if err := r.saveEvidence(ctx, transaction, caseFile.ID().UUID(), caseFile.ProjectID().UUID(), ev); err != nil {
			return fmt.Errorf("save evidence %s: %w", ev.ID(), err)
		}
	}

	if verdict, ok := caseFile.Verdict(); ok {
		if err := r.saveRulings(ctx, transaction, caseFile.ID().UUID(), verdict); err != nil {
			return fmt.Errorf("save rulings: %w", err)
		}
	}

	return transaction.Commit()
}

func (r *Repository) upsertCaseFile(ctx context.Context, transaction *database.Tx, caseFile model.CaseFile) error {
	_, err := transaction.ExecContext(ctx, `
		DELETE FROM casefiles WHERE project_id = ? AND commit_sha = ? AND id != ?
	`, caseFile.ProjectID().UUID(), caseFile.CommitSHA(), caseFile.ID().UUID())
	if err != nil {
		return fmt.Errorf("replace previous casefile: %w", err)
	}

	var verdictOutcome, verdictEvaluatedAt any
	if v, ok := caseFile.Verdict(); ok {
		verdictOutcome = v.Outcome().String()
		verdictEvaluatedAt = v.EvaluatedAt().UTC().Format(time.RFC3339)
	}

	_, err = transaction.ExecContext(ctx, `
		INSERT INTO casefiles (id, project_id, commit_sha, branch, started_at,
		                       verdict_outcome, verdict_evaluated_at,
		                       total_findings, new_findings, existing_findings, resolved_findings,
		                       is_fresh_evaluation)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0, ?)
		ON CONFLICT(id) DO UPDATE SET
		    verdict_outcome = excluded.verdict_outcome,
		    verdict_evaluated_at = excluded.verdict_evaluated_at,
		    total_findings = excluded.total_findings,
		    is_fresh_evaluation = excluded.is_fresh_evaluation
	`, caseFile.ID().UUID(), caseFile.ProjectID().UUID(), caseFile.CommitSHA(), caseFile.Branch(),
		caseFile.StartedAt().UTC().Format(time.RFC3339),
		verdictOutcome, verdictEvaluatedAt,
		countTotalFindings(caseFile),
		caseFile.IsFreshEvaluation())
	return err
}

func (r *Repository) saveEvidence(ctx context.Context, transaction *database.Tx, caseFileID, projectID uuid.UUID, evidenceItem evidence.Evidence) error {
	_, err := transaction.ExecContext(ctx, `
		INSERT INTO evidences (id, casefile_id, subtype, source, collected_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO NOTHING
	`, evidenceItem.ID().UUID(), caseFileID, evidenceItem.Subtype().String(), evidenceItem.Source(),
		evidenceItem.CollectedAt().UTC().Format(time.RFC3339))
	if err != nil {
		return err
	}

	return r.saveEvidenceContent(ctx, transaction, evidenceItem.ID().UUID(), caseFileID, projectID, evidenceItem.Content())
}

func (r *Repository) saveEvidenceContent(ctx context.Context, transaction *database.Tx, evidenceID, caseFileID, projectID uuid.UUID, content evidence.Content) error {
	switch typedContent := content.(type) {
	case finding.Content:
		return r.saveFindings(ctx, transaction, evidenceID, caseFileID, projectID, typedContent)
	case coverage.Content:
		return r.saveCoverage(ctx, transaction, evidenceID, typedContent)
	case coverage.PatchContent:
		return r.saveNewCodeCoverage(ctx, transaction, evidenceID, typedContent)
	case architecture.Content:
		return r.saveArchViolations(ctx, transaction, evidenceID, typedContent)
	case toolexecution.Content:
		return r.saveToolExecutions(ctx, transaction, evidenceID, typedContent)
	}
	return nil
}

func (r *Repository) saveFindings(ctx context.Context, transaction *database.Tx, evidenceID, caseFileID, projectID uuid.UUID, fc finding.Content) error {
	if _, err := transaction.ExecContext(ctx, "DELETE FROM findings WHERE evidence_id = ?", evidenceID); err != nil {
		return fmt.Errorf("clear existing findings: %w", err)
	}
	for _, fin := range fc.Findings() {
		_, err := transaction.ExecContext(ctx, `
			INSERT INTO findings (evidence_id, casefile_id, project_id,
			                                tool, rule_id, severity, file_path, line,
			                                message, fingerprint, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'new')
		`, evidenceID, caseFileID, projectID,
			fin.Tool(), fin.RuleID(), fin.Severity().String(),
			fin.FilePath(), fin.Line(), fin.Message(), fin.ID().Value())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) saveCoverage(ctx context.Context, transaction *database.Tx, evidenceID uuid.UUID, covContent coverage.Content) error {
	_, err := transaction.ExecContext(ctx, `
		INSERT INTO coverage_data (evidence_id, total_lines, covered_lines)
		VALUES (?, ?, ?)
		ON CONFLICT(evidence_id) DO NOTHING
	`, evidenceID, covContent.TotalLines(), covContent.CoveredLines())
	if err != nil {
		return err
	}

	if _, err := transaction.ExecContext(ctx, "DELETE FROM coverage_by_language WHERE evidence_id = ?", evidenceID); err != nil {
		return fmt.Errorf("clear existing coverage by language: %w", err)
	}
	for _, lc := range covContent.ByLanguage() {
		_, err := transaction.ExecContext(ctx, `
			INSERT INTO coverage_by_language (evidence_id, language, total_lines, covered_lines)
			VALUES (?, ?, ?, ?)
		`, evidenceID, lc.Language().String(), lc.TotalLines(), lc.CoveredLines())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) saveNewCodeCoverage(ctx context.Context, transaction *database.Tx, evidenceID uuid.UUID, ncc coverage.PatchContent) error {
	_, err := transaction.ExecContext(ctx, `
		INSERT INTO new_code_coverage_data (evidence_id, covered_lines, coverable_lines)
		VALUES (?, ?, ?)
		ON CONFLICT(evidence_id) DO NOTHING
	`, evidenceID, ncc.CoveredLines(), ncc.CoverableLines())
	return err
}

func (r *Repository) saveArchViolations(ctx context.Context, transaction *database.Tx, evidenceID uuid.UUID, ac architecture.Content) error {
	if _, err := transaction.ExecContext(ctx, "DELETE FROM architecture_violations WHERE evidence_id = ?", evidenceID); err != nil {
		return fmt.Errorf("clear existing violations: %w", err)
	}
	for _, v := range ac.Violations() {
		_, err := transaction.ExecContext(ctx, `
			INSERT INTO architecture_violations (evidence_id, rule, source_pkg, target_pkg, message)
			VALUES (?, ?, ?, ?, ?)
		`, evidenceID, v.Rule(), v.SourcePkg(), v.TargetPkg(), v.Message())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) saveToolExecutions(ctx context.Context, transaction *database.Tx, evidenceID uuid.UUID, tec toolexecution.Content) error {
	if _, err := transaction.ExecContext(ctx, "DELETE FROM tool_execution_failures WHERE evidence_id = ?", evidenceID); err != nil {
		return fmt.Errorf("clear existing tool execution failures: %w", err)
	}
	for _, failed := range tec.Failures() {
		_, err := transaction.ExecContext(ctx, `
			INSERT INTO tool_execution_failures (evidence_id, tool, reason)
			VALUES (?, ?, ?)
		`, evidenceID, failed.Tool(), failed.Reason())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) saveRulings(ctx context.Context, transaction *database.Tx, caseFileID uuid.UUID, verdict verdict.Result) error {
	_, err := transaction.ExecContext(ctx, "DELETE FROM rulings WHERE casefile_id = ?", caseFileID)
	if err != nil {
		return err
	}

	for i, ruling := range verdict.Rulings() {
		_, err := transaction.ExecContext(ctx, `
			INSERT INTO rulings (casefile_id, subtype, passed, detail, sort_order)
			VALUES (?, ?, ?, ?, ?)
		`, caseFileID, ruling.Subtype().String(),
			database.BoolToInt(ruling.Passed()), ruling.Detail(), i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) WriteCounters(ctx context.Context, caseFileID string, counters finalize.Counters) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE casefiles SET new_findings = ?, existing_findings = ?, resolved_findings = ?
		WHERE id = ?
	`, counters.NewCount, counters.ExistingCount, counters.ResolvedCount, caseFileID)
	return err
}

func (r *Repository) FindByID(ctx context.Context, caseFileID model.CaseFileID) (model.CaseFile, error) {
	var cfID, projectID uuid.UUID
	var commitSHA, branch, startedAtStr string
	var verdictOutcome, verdictEvaluatedAt sql.NullString
	var isFreshEval bool

	err := r.db.QueryRowContext(ctx, `
		SELECT id, project_id, commit_sha, branch, started_at,
		       verdict_outcome, verdict_evaluated_at, is_fresh_evaluation
		FROM casefiles WHERE id = ?
	`, caseFileID.UUID()).Scan(&cfID, &projectID, &commitSHA, &branch, &startedAtStr,
		&verdictOutcome, &verdictEvaluatedAt, &isFreshEval)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.CaseFile{}, fmt.Errorf("%w: %s", errCaseFileNotFound, caseFileID)
		}
		return model.CaseFile{}, fmt.Errorf("query casefile: %w", err)
	}

	startedAt, err := database.ParseTime(startedAtStr)
	if err != nil {
		return model.CaseFile{}, fmt.Errorf("parse started_at: %w", err)
	}

	evidences, err := r.loadEvidences(ctx, cfID)
	if err != nil {
		return model.CaseFile{}, fmt.Errorf("load evidences: %w", err)
	}

	var verdict *verdict.Result
	if verdictOutcome.Valid {
		v, err := r.loadVerdict(ctx, cfID, verdictOutcome.String, verdictEvaluatedAt.String)
		if err != nil {
			return model.CaseFile{}, fmt.Errorf("load verdict: %w", err)
		}
		verdict = &v
	}

	scannedCaseFileID := model.NewCaseFileID(cfID)
	projID := projectmodel.NewProjectID(projectID)

	return model.ReconstituteCaseFile(scannedCaseFileID, projID, commitSHA, branch, startedAt, evidences, verdict, isFreshEval)
}

func (r *Repository) loadEvidences(ctx context.Context, caseFileID uuid.UUID) ([]evidence.Evidence, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, subtype, source, collected_at
		FROM evidences WHERE casefile_id = ?
	`, caseFileID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	type row struct {
		id                           uuid.UUID
		subtype, source, collectedAt string
	}
	var scanned []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.subtype, &r.source, &r.collectedAt); err != nil {
			return nil, err
		}
		scanned = append(scanned, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var evidences []evidence.Evidence
	for _, s := range scanned {
		ev, err := r.reconstituteEvidence(ctx, s.id, s.subtype, s.source, s.collectedAt)
		if err != nil {
			return nil, err
		}
		if ev != nil {
			evidences = append(evidences, *ev)
		}
	}
	return evidences, nil
}

func (r *Repository) reconstituteEvidence(ctx context.Context, idVal uuid.UUID, subtypeStr, source, collectedAtStr string) (*evidence.Evidence, error) {
	evID := evidence.NewEvidenceID(idVal)
	subtype, err := evidence.NewSubtype(subtypeStr)
	if err != nil {
		return nil, err
	}
	collectedAt, err := database.ParseTime(collectedAtStr)
	if err != nil {
		return nil, err
	}

	content, err := r.loadEvidenceContent(ctx, idVal, subtype)
	if err != nil {
		return nil, fmt.Errorf("load content for evidence %s: %w", idVal, err)
	}
	if content == nil {
		return nil, nil
	}

	ev, err := evidence.ReconstituteEvidence(evID, subtype, source, content, collectedAt)
	if err != nil {
		return nil, err
	}
	return &ev, nil
}

func (r *Repository) loadEvidenceContent(ctx context.Context, evidenceID uuid.UUID, subtype evidence.Subtype) (evidence.Content, error) {
	if evidence.IsSubtypeFindingBased(subtype) {
		return r.loadFindingsContent(ctx, evidenceID, subtype)
	}
	switch subtype {
	case evidence.SubtypeCoverage:
		return r.loadCoverageContent(ctx, evidenceID)
	case evidence.SubtypeNewCodeCoverage:
		return r.loadNewCodeCoverageContent(ctx, evidenceID)
	case evidence.SubtypeArchitecture:
		return r.loadArchitectureContent(ctx, evidenceID)
	case evidence.SubtypeToolExecution:
		return r.loadToolExecutionContent(ctx, evidenceID)
	}
	return nil, nil
}

func (r *Repository) loadFindingsContent(ctx context.Context, evidenceID uuid.UUID, subtype evidence.Subtype) (evidence.Content, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tool, rule_id, severity, file_path, line, message, fingerprint
		FROM findings WHERE evidence_id = ?
	`, evidenceID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var findings []finding.Finding
	for rows.Next() {
		var tool, ruleID, severityStr, filePath, message, fpStr string
		var line int
		if err := rows.Scan(&tool, &ruleID, &severityStr, &filePath, &line, &message, &fpStr); err != nil {
			return nil, err
		}
		severity, err := finding.NewSeverity(severityStr)
		if err != nil {
			return nil, err
		}
		fp, err := finding.NewFingerprintID(fpStr)
		if err != nil {
			return nil, err
		}
		f, err := finding.NewFinding(tool, ruleID, severity, filePath, line, message, fp)
		if err != nil {
			return nil, err
		}
		findings = append(findings, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return finding.NewContent(subtype, findings)
}

func (r *Repository) loadCoverageContent(ctx context.Context, evidenceID uuid.UUID) (evidence.Content, error) {
	var totalLines, coveredLines int
	err := r.db.QueryRowContext(ctx, `
		SELECT total_lines, covered_lines FROM coverage_data WHERE evidence_id = ?
	`, evidenceID).Scan(&totalLines, &coveredLines)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	langRows, err := r.db.QueryContext(ctx, `
		SELECT language, total_lines, covered_lines
		FROM coverage_by_language WHERE evidence_id = ?
	`, evidenceID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = langRows.Close() }()

	var byLanguage []coverage.LanguageStats
	for langRows.Next() {
		var langStr string
		var langTotal, langCovered int
		if err := langRows.Scan(&langStr, &langTotal, &langCovered); err != nil {
			return nil, err
		}
		lang, err := coverage.NewLanguage(langStr)
		if err != nil {
			return nil, err
		}
		lCov, err := coverage.NewLanguageStats(lang, langTotal, langCovered)
		if err != nil {
			return nil, err
		}
		byLanguage = append(byLanguage, lCov)
	}
	if err := langRows.Err(); err != nil {
		return nil, err
	}

	return coverage.NewContent(totalLines, coveredLines, byLanguage)
}

func (r *Repository) loadNewCodeCoverageContent(ctx context.Context, evidenceID uuid.UUID) (evidence.Content, error) {
	var coveredLines, coverableLines int
	err := r.db.QueryRowContext(ctx, `
		SELECT covered_lines, coverable_lines FROM new_code_coverage_data WHERE evidence_id = ?
	`, evidenceID).Scan(&coveredLines, &coverableLines)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return coverage.NewPatchContent(coveredLines, coverableLines)
}

func (r *Repository) loadArchitectureContent(ctx context.Context, evidenceID uuid.UUID) (evidence.Content, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT rule, source_pkg, target_pkg, message
		FROM architecture_violations WHERE evidence_id = ?
	`, evidenceID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var violations []architecture.Violation
	for rows.Next() {
		var rule, sourcePkg, targetPkg, message string
		if err := rows.Scan(&rule, &sourcePkg, &targetPkg, &message); err != nil {
			return nil, err
		}
		v, err := architecture.NewViolation(rule, sourcePkg, targetPkg, message)
		if err != nil {
			return nil, err
		}
		violations = append(violations, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return architecture.NewContent(violations)
}

func (r *Repository) loadToolExecutionContent(ctx context.Context, evidenceID uuid.UUID) (evidence.Content, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tool, reason
		FROM tool_execution_failures WHERE evidence_id = ?
	`, evidenceID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var failures []toolexecution.Failure
	for rows.Next() {
		var tool, reason string
		if err := rows.Scan(&tool, &reason); err != nil {
			return nil, err
		}
		failed, err := toolexecution.NewFailure(tool, reason)
		if err != nil {
			return nil, err
		}
		failures = append(failures, failed)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return toolexecution.NewContent(failures)
}

func (r *Repository) loadVerdict(ctx context.Context, caseFileID uuid.UUID, outcomeStr, evaluatedAtStr string) (verdict.Result, error) {
	evaluatedAt, err := database.ParseTime(evaluatedAtStr)
	if err != nil {
		return verdict.Result{}, fmt.Errorf("parse verdict evaluated_at: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT subtype, passed, detail
		FROM rulings WHERE casefile_id = ? ORDER BY sort_order
	`, caseFileID)
	if err != nil {
		return verdict.Result{}, err
	}
	defer func() { _ = rows.Close() }()

	var rulings []verdict.Ruling
	for rows.Next() {
		var subtypeStr, detail string
		var passedInt int
		if err := rows.Scan(&subtypeStr, &passedInt, &detail); err != nil {
			return verdict.Result{}, err
		}
		subtype, err := evidence.NewSubtype(subtypeStr)
		if err != nil {
			return verdict.Result{}, err
		}
		rulings = append(rulings, verdict.NewRuling(subtype, passedInt != 0, detail))
	}
	if err := rows.Err(); err != nil {
		return verdict.Result{}, err
	}

	return verdict.ReconstituteResult(outcomeStr, rulings, evaluatedAt)
}

func (r *Repository) FindByProject(ctx context.Context, projectID projectmodel.ProjectID) ([]model.CaseFile, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, commit_sha, branch, started_at, is_fresh_evaluation
		FROM casefiles WHERE project_id = ?
		ORDER BY started_at DESC
	`, projectID.UUID())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	type row struct {
		id, projectID                uuid.UUID
		commitSHA, branch, startedAt string
		isFreshEval                  bool
	}
	var scanned []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.projectID, &r.commitSHA, &r.branch, &r.startedAt, &r.isFreshEval); err != nil {
			return nil, err
		}
		scanned = append(scanned, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var result []model.CaseFile
	for _, row := range scanned {
		startedAt, err := database.ParseTime(row.startedAt)
		if err != nil {
			return nil, err
		}
		caseFileID := model.NewCaseFileID(row.id)
		projID := projectmodel.NewProjectID(row.projectID)
		cf, err := model.ReconstituteCaseFile(caseFileID, projID, row.commitSHA, row.branch, startedAt, nil, nil, row.isFreshEval)
		if err != nil {
			return nil, err
		}
		result = append(result, cf)
	}
	return result, nil
}

func (r *Repository) FindLatestByBranch(ctx context.Context, projectID projectmodel.ProjectID, branch string) (model.CaseFile, error) {
	var cfID uuid.UUID
	err := r.db.QueryRowContext(ctx, `
		SELECT id FROM casefiles
		WHERE project_id = ? AND branch = ?
		ORDER BY started_at DESC LIMIT 1
	`, projectID.UUID(), branch).Scan(&cfID)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.CaseFile{}, fmt.Errorf("%w: project %s branch %s", errCaseFileNotFound, projectID, branch)
		}
		return model.CaseFile{}, fmt.Errorf("query latest casefile: %w", err)
	}

	id := model.NewCaseFileID(cfID)
	return r.FindByID(ctx, id)
}

func (r *Repository) FindFingerprintIDsByBranch(ctx context.Context, projectID projectmodel.ProjectID, branch string) ([]finding.FingerprintID, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT f.fingerprint
		FROM findings f
		JOIN casefiles cf ON f.casefile_id = cf.id
		WHERE cf.project_id = ? AND cf.branch = ?
		  AND cf.id = (
		      SELECT id FROM casefiles
		      WHERE project_id = ? AND branch = ?
		      ORDER BY started_at DESC LIMIT 1
		  )
		  AND f.status != 'resolved'
	`, projectID.UUID(), branch, projectID.UUID(), branch)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var fingerprints []finding.FingerprintID
	for rows.Next() {
		var fpStr string
		if err := rows.Scan(&fpStr); err != nil {
			return nil, err
		}
		fp, err := finding.NewFingerprintID(fpStr)
		if err != nil {
			return nil, err
		}
		fingerprints = append(fingerprints, fp)
	}
	return fingerprints, rows.Err()
}

func countTotalFindings(cf model.CaseFile) int {
	seen := make(map[string]struct{})
	for _, ev := range cf.Evidences() {
		fc, ok := ev.Content().(finding.Content)
		if !ok {
			continue
		}
		for _, f := range fc.Findings() {
			seen[f.ID().Value()] = struct{}{}
		}
	}
	return len(seen)
}
