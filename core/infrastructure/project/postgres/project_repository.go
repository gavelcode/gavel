package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
	"github.com/usegavel/gavel/core/domain/shared/failure"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
)

var errProjectNotFound = failure.New("project not found", failure.NotFound)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(ctx context.Context, project model.Project) error {
	transaction, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	_, err = transaction.ExecContext(ctx, `
		INSERT INTO projects (id, key, name, target_pattern, default_branch, tenant_id)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
		    name = excluded.name,
		    target_pattern = excluded.target_pattern,
		    default_branch = excluded.default_branch,
		    updated_at = ?
	`, project.ID().UUID(), project.Key(), project.Name(),
		project.TargetPattern(), project.DefaultBranch(), project.TenantID().UUID(), database.Now())
	if err != nil {
		return fmt.Errorf("upsert project: %w", err)
	}

	if err := r.replaceLanguages(ctx, transaction, project); err != nil {
		return fmt.Errorf("replace languages: %w", err)
	}

	if err := r.replaceQualityGateRules(ctx, transaction, project); err != nil {
		return fmt.Errorf("replace quality gate rules: %w", err)
	}

	if err := r.replaceBaselines(ctx, transaction, project); err != nil {
		return fmt.Errorf("replace baselines: %w", err)
	}

	return transaction.Commit()
}

func (r *Repository) replaceLanguages(ctx context.Context, transaction *database.Tx, project model.Project) error {
	_, err := transaction.ExecContext(ctx, "DELETE FROM project_languages WHERE project_id = ?", project.ID().UUID())
	if err != nil {
		return err
	}

	for _, lang := range project.Languages() {
		_, err = transaction.ExecContext(ctx, `
			INSERT INTO project_languages (project_id, language) VALUES (?, ?)
		`, project.ID().UUID(), lang.String())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) replaceQualityGateRules(ctx context.Context, transaction *database.Tx, project model.Project) error {
	_, err := transaction.ExecContext(ctx, "DELETE FROM project_quality_gate_rules WHERE project_id = ?", project.ID().UUID())
	if err != nil {
		return err
	}

	for sortOrder, rule := range project.Gate().Rules() {
		strategyType, strategyParams := serializeStrategy(rule.Strategy())
		_, err = transaction.ExecContext(ctx, `
			INSERT INTO project_quality_gate_rules
			    (project_id, subtype, strategy_type, strategy_params, sort_order)
			VALUES (?, ?, ?, ?, ?)
		`, project.ID().UUID(), rule.Subtype().String(),
			strategyType, strategyParams, sortOrder)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, tenantID tenant.TenantID, id model.ProjectID) (model.Project, error) {
	return r.findProject(ctx, "SELECT id, key, name, target_pattern, default_branch FROM projects WHERE id = ? AND tenant_id = ?", tenantID, id.UUID(), id.String())
}

func (r *Repository) FindByName(ctx context.Context, tenantID tenant.TenantID, name string) (model.Project, error) {
	return r.findProject(ctx, "SELECT id, key, name, target_pattern, default_branch FROM projects WHERE name = ? AND tenant_id = ?", tenantID, name, name)
}

func (r *Repository) FindByKey(ctx context.Context, tenantID tenant.TenantID, key string) (model.Project, error) {
	return r.findProject(ctx, "SELECT id, key, name, target_pattern, default_branch FROM projects WHERE key = ? AND tenant_id = ?", tenantID, key, key)
}

func (r *Repository) findProject(ctx context.Context, query string, tenantID tenant.TenantID, param any, label string) (model.Project, error) {
	var idVal uuid.UUID
	var key, name, targetPattern, branch string
	err := r.db.QueryRowContext(ctx, query, param, tenantID.UUID()).Scan(&idVal, &key, &name, &targetPattern, &branch)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Project{}, fmt.Errorf("%w: %s", errProjectNotFound, label)
		}
		return model.Project{}, fmt.Errorf("query project: %w", err)
	}

	projectID := model.NewProjectID(idVal)

	languages, err := r.loadLanguages(ctx, idVal)
	if err != nil {
		return model.Project{}, fmt.Errorf("load languages: %w", err)
	}

	qualityGate, err := r.loadQualityGate(ctx, idVal)
	if err != nil {
		return model.Project{}, fmt.Errorf("load quality gate: %w", err)
	}

	baselines, err := r.loadBaselines(ctx, idVal)
	if err != nil {
		return model.Project{}, fmt.Errorf("load baselines: %w", err)
	}

	return model.ReconstituteProject(projectID, tenantID, key, name, targetPattern, branch, languages, qualityGate, nil, baselines)
}

func (r *Repository) loadLanguages(ctx context.Context, projectID uuid.UUID) ([]coverage.Language, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT language FROM project_languages WHERE project_id = ?
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var languages []coverage.Language
	for rows.Next() {
		var langStr string
		if err := rows.Scan(&langStr); err != nil {
			return nil, err
		}
		lang, err := coverage.NewLanguage(langStr)
		if err != nil {
			return nil, err
		}
		languages = append(languages, lang)
	}
	return languages, rows.Err()
}

func (r *Repository) loadQualityGate(ctx context.Context, projectID uuid.UUID) (qualitygate.Gate, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT subtype, strategy_type, strategy_params
		FROM project_quality_gate_rules
		WHERE project_id = ? ORDER BY sort_order
	`, projectID)
	if err != nil {
		return qualitygate.Gate{}, err
	}
	defer func() { _ = rows.Close() }()

	var rules []qualitygate.Rule
	for rows.Next() {
		var subtypeStr, strategyType, strategyParams string
		if err := rows.Scan(&subtypeStr, &strategyType, &strategyParams); err != nil {
			return qualitygate.Gate{}, err
		}

		subtype, err := evidence.NewSubtype(subtypeStr)
		if err != nil {
			return qualitygate.Gate{}, err
		}
		strategy, err := deserializeStrategy(strategyType, strategyParams)
		if err != nil {
			return qualitygate.Gate{}, fmt.Errorf("deserialize strategy %q: %w", strategyType, err)
		}
		rule, err := qualitygate.NewRule(subtype, strategy)
		if err != nil {
			return qualitygate.Gate{}, err
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return qualitygate.Gate{}, err
	}

	return qualitygate.NewGate(rules)
}

func (r *Repository) replaceBaselines(ctx context.Context, transaction *database.Tx, project model.Project) error {
	_, err := transaction.ExecContext(ctx, "DELETE FROM project_baselines WHERE project_id = ?", project.ID().UUID())
	if err != nil {
		return err
	}

	for branch, baseline := range project.Baselines() {
		fps, err := json.Marshal(baseline.Fingerprints())
		if err != nil {
			return fmt.Errorf("marshal fingerprints: %w", err)
		}
		archIDs, err := json.Marshal(baseline.ArchIDs())
		if err != nil {
			return fmt.Errorf("marshal arch ids: %w", err)
		}
		var covPercent *float64
		if cp := baseline.CoveragePercent(); cp != nil {
			covPercent = cp
		}
		var covByFile *string
		if fc := baseline.FileCoverage(); len(fc) > 0 {
			fcJSON, err := marshalFileCoverage(fc)
			if err != nil {
				return fmt.Errorf("marshal file coverage: %w", err)
			}
			s := string(fcJSON)
			covByFile = &s
		}
		_, err = transaction.ExecContext(ctx, `
			INSERT INTO project_baselines (project_id, branch, fingerprints, arch_ids, coverage_percent, coverage_by_file)
			VALUES (?, ?, ?, ?, ?, ?)
		`, project.ID().UUID(), branch, string(fps), string(archIDs), covPercent, covByFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) loadBaselines(ctx context.Context, projectID uuid.UUID) (map[string]model.Baseline, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT branch, fingerprints, arch_ids, coverage_percent, coverage_by_file
		FROM project_baselines WHERE project_id = ?
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	baselines := make(map[string]model.Baseline)
	for rows.Next() {
		var branch, fpsJSON, archJSON string
		var covPercent *float64
		var covByFileJSON *string
		if err := rows.Scan(&branch, &fpsJSON, &archJSON, &covPercent, &covByFileJSON); err != nil {
			return nil, err
		}
		var fps []string
		if err := json.Unmarshal([]byte(fpsJSON), &fps); err != nil {
			return nil, fmt.Errorf("unmarshal fingerprints: %w", err)
		}
		var archIDs []string
		if err := json.Unmarshal([]byte(archJSON), &archIDs); err != nil {
			return nil, fmt.Errorf("unmarshal arch ids: %w", err)
		}
		var fileCov []model.FileCoverageEntry
		if covByFileJSON != nil {
			fileCov, err = unmarshalFileCoverage([]byte(*covByFileJSON))
			if err != nil {
				return nil, fmt.Errorf("unmarshal file coverage: %w", err)
			}
		}
		baselines[branch] = model.NewBaseline(fps, archIDs, covPercent, fileCov)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(baselines) == 0 {
		return nil, nil
	}
	return baselines, nil
}

type fileCoverageDTO struct {
	FilePath  string `json:"file_path"`
	Covered   []int  `json:"covered"`
	Uncovered []int  `json:"uncovered,omitempty"`
}

func marshalFileCoverage(entries []model.FileCoverageEntry) ([]byte, error) {
	dtos := make([]fileCoverageDTO, 0, len(entries))
	for _, e := range entries {
		dtos = append(dtos, fileCoverageDTO{
			FilePath:  e.FilePath(),
			Covered:   e.Covered(),
			Uncovered: e.Uncovered(),
		})
	}
	return json.Marshal(dtos)
}

func unmarshalFileCoverage(data []byte) ([]model.FileCoverageEntry, error) {
	var dtos []fileCoverageDTO
	if err := json.Unmarshal(data, &dtos); err != nil {
		return nil, err
	}
	entries := make([]model.FileCoverageEntry, 0, len(dtos))
	for _, d := range dtos {
		entry, err := model.NewFileCoverageEntry(d.FilePath, d.Covered, d.Uncovered)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func serializeStrategy(s qualitygate.Strategy) (string, string) {
	switch strategy := s.(type) {
	case qualitygate.ZeroTolerance:
		return "zero_tolerance", "{}"
	case qualitygate.CountBySeverity:
		params, _ := json.Marshal(map[string]int{
			"maxError": strategy.MaxError(), "maxWarning": strategy.MaxWarning(), "maxNote": strategy.MaxNote(),
		})
		return "count_by_severity", string(params)
	case qualitygate.MinPercentage:
		params, _ := json.Marshal(map[string]float64{"min": strategy.Min()})
		return "min_percentage", string(params)
	case qualitygate.ForbiddenList:
		params, _ := json.Marshal(map[string][]string{"forbidden": strategy.Forbidden()})
		return "forbidden_list", string(params)
	case qualitygate.MaxViolations:
		params, _ := json.Marshal(map[string]int{"max": strategy.Max()})
		return "max_violations", string(params)
	case qualitygate.MinNewCodeCoverage:
		params, _ := json.Marshal(map[string]float64{"min": strategy.Min()})
		return "min_new_code_coverage", string(params)
	}
	return "unknown", "{}"
}

func deserializeStrategy(strategyType, paramsJSON string) (qualitygate.Strategy, error) {
	switch strategyType {
	case "zero_tolerance":
		return qualitygate.NewZeroTolerance(), nil

	case "count_by_severity":
		var params struct {
			MaxError   int `json:"maxError"`
			MaxWarning int `json:"maxWarning"`
			MaxNote    int `json:"maxNote"`
		}
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return nil, fmt.Errorf("unmarshal count_by_severity params: %w", err)
		}
		return qualitygate.NewCountBySeverity(params.MaxError, params.MaxWarning, params.MaxNote)

	case "min_percentage":
		var params struct {
			Min float64 `json:"min"`
		}
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return nil, fmt.Errorf("unmarshal min_percentage params: %w", err)
		}
		return qualitygate.NewMinPercentage(params.Min)

	case "forbidden_list":
		var params struct {
			Forbidden []string `json:"forbidden"`
		}
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return nil, fmt.Errorf("unmarshal forbidden_list params: %w", err)
		}
		return qualitygate.NewForbiddenList(params.Forbidden)

	case "max_violations":
		var params struct {
			Max int `json:"max"`
		}
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return nil, fmt.Errorf("unmarshal max_violations params: %w", err)
		}
		return qualitygate.NewMaxViolations(params.Max)

	case "min_new_code_coverage":
		var params struct {
			Min float64 `json:"min"`
		}
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return nil, fmt.Errorf("unmarshal min_new_code_coverage params: %w", err)
		}
		return qualitygate.NewMinNewCodeCoverage(params.Min)
	}

	return nil, fmt.Errorf("unknown strategy type %q", strategyType)
}
