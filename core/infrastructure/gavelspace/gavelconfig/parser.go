package gavelconfig

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	gavelspacemodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

func ParseFile(path string) (WorkspaceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return WorkspaceConfig{}, fmt.Errorf("%w: %s: %w", ErrReadConfig, path, err)
	}
	return Parse(data)
}

func Parse(data []byte) (WorkspaceConfig, error) {
	var dto configDTO
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&dto); err != nil {
		return WorkspaceConfig{}, fmt.Errorf("%w: %w", ErrParseConfig, err)
	}
	cfg, err := mapToDomain(dto)
	if err != nil {

		return WorkspaceConfig{}, fmt.Errorf("%w: %w", ErrParseConfig, err)
	}
	return cfg, nil
}

func mapToDomain(dto configDTO) (WorkspaceConfig, error) {
	if len(dto.Projects) == 0 {
		return WorkspaceConfig{}, fmt.Errorf("%w: must define at least one project", ErrParseConfig)
	}

	gavelspace, err := gavelspacemodel.NewGavelspace(tenant.LocalTenantID, dto.Name)
	if err != nil {
		return WorkspaceConfig{}, fmt.Errorf("name: %w", err)
	}

	now := time.Now().UTC()
	projects := make([]projectmodel.Project, 0, len(dto.Projects))
	covOptions := make(map[string]CoverageOptions, len(dto.Projects))

	for index, pdto := range dto.Projects {
		project, err := buildProject(pdto, now)
		if err != nil {
			return WorkspaceConfig{}, fmt.Errorf("projects[%d] %q: %w", index, pdto.Name, err)
		}

		ref, err := gavelspacemodel.NewProjectRef(project.ID(), pdto.Pattern)
		if err != nil {
			return WorkspaceConfig{}, fmt.Errorf("projects[%d] %q: project ref: %w", index, pdto.Name, err)
		}
		if err := gavelspace.AddProject(ref, now); err != nil {
			return WorkspaceConfig{}, fmt.Errorf("projects[%d] %q: %w", index, pdto.Name, err)
		}

		if pdto.CoverageOptions != nil {
			covOptions[pdto.Name] = CoverageOptions{
				testSizeFilters:       pdto.CoverageOptions.TestSizeFilters,
				testTagFilters:        pdto.CoverageOptions.TestTagFilters,
				instrumentationFilter: pdto.CoverageOptions.InstrumentationFilter,
			}
		}

		projects = append(projects, project)
	}

	if err := validateFindingsSource(dto.FindingsSource); err != nil {
		return WorkspaceConfig{}, err
	}

	gavelspace.SetServerConfig(gavelspacemodel.NewServerConfig(dto.Server.URL, dto.Server.Token))
	gavelspace.SetFindingsSource(dto.FindingsSource)

	return WorkspaceConfig{
		gavelspace:      gavelspace,
		projects:        projects,
		coverageOptions: covOptions,
		server:          ServerConfig{url: dto.Server.URL, token: dto.Server.Token},
		findingsSource:  dto.FindingsSource,
	}, nil
}

func buildProject(dto projectDTO, now time.Time) (projectmodel.Project, error) {
	key := slugify(dto.Name)
	project, err := projectmodel.NewProject(key, dto.Name, dto.Pattern)
	if err != nil {
		return projectmodel.Project{}, err
	}

	languages, selection, err := buildTooling(dto.Tooling)
	if err != nil {
		return projectmodel.Project{}, fmt.Errorf("tooling: %w", err)
	}
	if len(languages) > 0 {
		project.UpdateLanguages(languages, now)
		project.UpdateToolSelection(selection, now)
	}

	qg, err := buildQualityGate(dto.Gate)
	if err != nil {
		return projectmodel.Project{}, err
	}
	if len(qg.Rules()) > 0 {
		project.UpdateQualityGate(qg, now)
	}

	if len(dto.Exclude) > 0 {
		if err := project.UpdateExcludePatterns(dto.Exclude, now); err != nil {
			return projectmodel.Project{}, fmt.Errorf("exclude: %w", err)
		}
	}

	return project, nil
}

func buildTooling(tooling map[string][]string) ([]coverage.Language, map[string][]string, error) {
	names := make([]string, 0, len(tooling))
	for name := range tooling {
		names = append(names, name)
	}
	sort.Strings(names)

	languages := make([]coverage.Language, 0, len(names))
	selection := make(map[string][]string, len(names))
	for _, name := range names {
		tools := tooling[name]
		if len(tools) == 0 {
			return nil, nil, fmt.Errorf("language %q must list at least one tool", name)
		}
		lang, err := coverage.NewLanguage(name)
		if err != nil {
			return nil, nil, err
		}
		languages = append(languages, lang)
		selection[name] = tools
	}
	return languages, selection, nil
}

func buildQualityGate(dto qualityGateDTO) (qualitygate.Gate, error) {
	var rules []qualitygate.Rule

	if dto.Findings != nil {
		rule, err := buildFindingsRule(*dto.Findings)
		if err != nil {
			return qualitygate.Gate{}, fmt.Errorf("findings: %w", err)
		}
		rules = append(rules, rule)
	}

	if dto.Coverage != nil {
		rule, err := buildCoverageRule(*dto.Coverage)
		if err != nil {
			return qualitygate.Gate{}, fmt.Errorf("coverage: %w", err)
		}
		rules = append(rules, rule)
	}

	if dto.NewCodeCoverage != nil {
		rule, err := buildNewCodeCoverageRule(*dto.NewCodeCoverage)
		if err != nil {
			return qualitygate.Gate{}, fmt.Errorf("new_code_coverage: %w", err)
		}
		rules = append(rules, rule)
	}

	if dto.Violations != nil {
		rule, err := buildViolationsRule(*dto.Violations)
		if err != nil {
			return qualitygate.Gate{}, fmt.Errorf("violations: %w", err)
		}
		rules = append(rules, rule)
	}

	if dto.License != nil {
		rule, err := buildLicenseRule(*dto.License)
		if err != nil {
			return qualitygate.Gate{}, fmt.Errorf("license: %w", err)
		}
		rules = append(rules, rule)
	}

	if len(rules) == 0 {
		return qualitygate.Gate{}, nil
	}
	return qualitygate.NewGate(rules)
}

func buildFindingsRule(dto findingsRuleDTO) (qualitygate.Rule, error) {
	maxError := valueOrZero(dto.MaxError)
	maxWarning := valueOrZero(dto.MaxWarning)
	maxNote := valueOrZero(dto.MaxNote)

	var strategy qualitygate.Strategy
	if maxError == 0 && maxWarning == 0 && maxNote == 0 &&
		dto.MaxError == nil && dto.MaxWarning == nil && dto.MaxNote == nil {
		strategy = qualitygate.NewZeroTolerance()
	} else {
		s, err := qualitygate.NewCountBySeverity(maxError, maxWarning, maxNote)
		if err != nil {
			return qualitygate.Rule{}, err
		}
		strategy = s
	}
	var opts []qualitygate.RuleOption
	if dto.MinResolved != nil {
		opts = append(opts, qualitygate.WithMinResolved(*dto.MinResolved))
	}
	return qualitygate.NewRule(
		evidence.SubtypeCodeQuality,
		strategy,
		opts...,
	)
}

func buildCoverageRule(dto coverageRuleDTO) (qualitygate.Rule, error) {
	strategy, err := qualitygate.NewMinPercentage(dto.Min)
	if err != nil {
		return qualitygate.Rule{}, err
	}
	var opts []qualitygate.RuleOption
	if dto.MinDelta != nil {
		opts = append(opts, qualitygate.WithMinDelta(*dto.MinDelta))
	}
	return qualitygate.NewRule(
		evidence.SubtypeCoverage,
		strategy,
		opts...,
	)
}

func buildNewCodeCoverageRule(dto newCodeCoverageRuleDTO) (qualitygate.Rule, error) {
	strategy, err := qualitygate.NewMinNewCodeCoverage(dto.Min)
	if err != nil {
		return qualitygate.Rule{}, err
	}
	return qualitygate.NewRule(
		evidence.SubtypeNewCodeCoverage,
		strategy,
	)
}

func buildViolationsRule(dto violationsRuleDTO) (qualitygate.Rule, error) {
	strategy, err := qualitygate.NewMaxViolations(dto.Max)
	if err != nil {
		return qualitygate.Rule{}, err
	}
	var opts []qualitygate.RuleOption
	if dto.MinResolved != nil {
		opts = append(opts, qualitygate.WithMinResolved(*dto.MinResolved))
	}
	return qualitygate.NewRule(
		evidence.SubtypeArchitecture,
		strategy,
		opts...,
	)
}

func buildLicenseRule(dto licenseRuleDTO) (qualitygate.Rule, error) {
	strategy, err := qualitygate.NewForbiddenList(dto.Forbidden)
	if err != nil {
		return qualitygate.Rule{}, err
	}
	return qualitygate.NewRule(
		evidence.SubtypeLicense,
		strategy,
	)
}

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	var builder strings.Builder
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			builder.WriteRune(c)
		}
	}
	return strings.Trim(builder.String(), "-")
}

func valueOrZero(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

var validFindingsSources = map[string]bool{
	"":           true,
	"auto":       true,
	"gavel":      true,
	"rules_lint": true,
}

func validateFindingsSource(source string) error {
	if validFindingsSources[source] {
		return nil
	}

	return fmt.Errorf("findings_source: unknown value %q (valid: auto, gavel, rules_lint)", source)
}
