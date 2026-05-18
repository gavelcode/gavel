package gavelconfig_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
	"github.com/usegavel/gavel/core/domain/shared/failure"
	"github.com/usegavel/gavel/core/infrastructure/gavelspace/gavelconfig"
)

func locateGavelYaml(t *testing.T) string {
	t.Helper()
	if path, err := runfiles.Rlocation("_main/.gavel/gavel.yaml"); err == nil {
		return path
	}
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		candidate := filepath.Join(dir, ".gavel", "gavel.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal(".gavel/gavel.yaml not found in runfiles or any ancestor directory")
		}
		dir = parent
	}
}

func TestParseGavelProjectYaml(t *testing.T) {
	path := locateGavelYaml(t)

	config, err := gavelconfig.ParseFile(path)
	require.NoError(t, err)

	gs := config.Gavelspace()
	assert.Equal(t, "gavel", gs.ID().String())

	projects := config.Projects()
	require.Len(t, projects, 5, "gavel-project's gavel.yaml must declare 5 projects: core, cli, server, web, tools")

	expected := map[string]string{
		"core":   "//core/...",
		"cli":    "//apps/cli/...",
		"server": "//apps/server/...",
		"web":    "//apps/web/...",
		"tools":  "//tools/...",
	}
	for _, p := range projects {
		pattern, ok := expected[p.Name()]
		assert.True(t, ok, "unexpected project %q in gavel.yaml", p.Name())
		assert.Equal(t, pattern, p.TargetPattern(), "target pattern for project %q", p.Name())
	}
}

func TestParseShouldProduceGavelspaceWithProjects(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    tooling: [java, kotlin]
  - name: frontend
    pattern: //web/...
    tooling: [typescript]
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	gs := config.Gavelspace()
	assert.Equal(t, "my-monorepo", gs.ID().String())
	assert.Len(t, gs.Projects(), 2)
	assert.Len(t, config.Projects(), 2)

	backend := config.Projects()[0]
	assert.Equal(t, "backend", backend.Name())
	assert.Equal(t, "//backend/...", backend.TargetPattern())
	assert.Len(t, backend.Languages(), 2)
	assert.Equal(t, "java", backend.Languages()[0].String())
	assert.Equal(t, "kotlin", backend.Languages()[1].String())

	frontend := config.Projects()[1]
	assert.Equal(t, "frontend", frontend.Name())
	assert.Equal(t, "//web/...", frontend.TargetPattern())
	assert.Len(t, frontend.Languages(), 1)
	assert.Equal(t, "typescript", frontend.Languages()[0].String())
}

func TestParseShouldProduceQualityGateWithZeroTolerance(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      findings: {}
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	qg := config.Projects()[0].Gate()
	rules := qg.Rules()
	require.Len(t, rules, 1)

	rule := rules[0]
	assert.Equal(t, evidence.SubtypeCodeQuality, rule.Subtype())
	assert.IsType(t, qualitygate.ZeroTolerance{}, rule.Strategy())
}

func TestParseShouldProduceQualityGateWithCountBySeverity(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      findings:
        max_error: 0
        max_warning: 10
        max_note: 50
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	qg := config.Projects()[0].Gate()
	rules := qg.Rules()
	require.Len(t, rules, 1)

	rule := rules[0]
	assert.Equal(t, evidence.SubtypeCodeQuality, rule.Subtype())
	assert.IsType(t, qualitygate.CountBySeverity{}, rule.Strategy())
}

func TestParseShouldProduceQualityGateWithMinCoverage(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      coverage:
        min: 80
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	qg := config.Projects()[0].Gate()
	rules := qg.Rules()
	require.Len(t, rules, 1)

	rule := rules[0]
	assert.Equal(t, evidence.SubtypeCoverage, rule.Subtype())
	assert.IsType(t, qualitygate.MinPercentage{}, rule.Strategy())
}

func TestParseShouldProduceQualityGateWithAllRules(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      findings:
        max_error: 0
      coverage:
        min: 80
      new_code_coverage:
        min: 80
      architecture_violations:
        max: 0
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	qg := config.Projects()[0].Gate()
	assert.Len(t, qg.Rules(), 4)
}

func TestParseShouldProduceQualityGateWithMaxViolations(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      architecture_violations:
        max: 0
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	qg := config.Projects()[0].Gate()
	rules := qg.Rules()
	require.Len(t, rules, 1)

	rule := rules[0]
	assert.Equal(t, evidence.SubtypeArchitecture, rule.Subtype())
	assert.IsType(t, qualitygate.MaxViolations{}, rule.Strategy())
}

func TestParseShouldProduceQualityGateWithForbiddenLicenseList(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      license:
        forbidden:
          - GPL-3.0
          - AGPL-3.0
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	qg := config.Projects()[0].Gate()
	rules := qg.Rules()
	require.Len(t, rules, 1)

	rule := rules[0]
	assert.Equal(t, evidence.SubtypeLicense, rule.Subtype())
	require.IsType(t, qualitygate.ForbiddenList{}, rule.Strategy())
	strategy := rule.Strategy().(qualitygate.ForbiddenList)
	assert.ElementsMatch(t, []string{"GPL-3.0", "AGPL-3.0"}, strategy.Forbidden())
}

func TestParseShouldRejectLicenseRuleWithEmptyForbiddenList(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      license:
        forbidden: []
`)

	_, err := gavelconfig.Parse(data)
	assert.Error(t, err, "license rule with empty forbidden list must be rejected")
}


func TestParseShouldProduceFindingsRuleWithMinResolved(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      findings:
        max_error: 0
        min_resolved: 5
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	qg := config.Projects()[0].Gate()
	rules := qg.Rules()
	require.Len(t, rules, 1)

	rule := rules[0]
	require.NotNil(t, rule.MinResolved())
	assert.Equal(t, 5, *rule.MinResolved())
}

func TestParseShouldProduceViolationsRuleWithMinResolved(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      architecture_violations:
        max: 0
        min_resolved: 3
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	qg := config.Projects()[0].Gate()
	rules := qg.Rules()
	require.Len(t, rules, 1)

	rule := rules[0]
	require.NotNil(t, rule.MinResolved())
	assert.Equal(t, 3, *rule.MinResolved())
}

func TestParseShouldProduceCoverageRuleWithMinDelta(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      coverage:
        min: 80
        min_delta: 0
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	qg := config.Projects()[0].Gate()
	rules := qg.Rules()
	require.Len(t, rules, 1)

	rule := rules[0]
	require.NotNil(t, rule.MinDelta())
	assert.InDelta(t, 0.0, *rule.MinDelta(), 0.001)
}

func TestParseShouldRejectNegativeMinResolved(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      findings:
        max_error: 0
        min_resolved: -1
`)

	_, err := gavelconfig.Parse(data)
	assert.Error(t, err)
}

func TestParseShouldOmitDeltaFieldsWhenNotSpecified(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      findings:
        max_error: 0
      coverage:
        min: 80
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	qg := config.Projects()[0].Gate()
	rules := qg.Rules()
	require.Len(t, rules, 2)

	for _, rule := range rules {
		assert.Nil(t, rule.MinResolved(), "no min_resolved specified for %s", rule.Subtype())
		assert.Nil(t, rule.MinDelta(), "no min_delta specified for %s", rule.Subtype())
	}
}

func TestParseShouldRejectNegativeMaxViolations(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      architecture_violations:
        max: -1
`)

	_, err := gavelconfig.Parse(data)
	assert.Error(t, err)
}

func TestParseShouldProduceServerConfig(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
server:
  url: https://gavel.example.com
  token: gav_xxx
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	assert.Equal(t, "https://gavel.example.com", config.Server().URL())
	assert.Equal(t, "gav_xxx", config.Server().Token())
	assert.False(t, config.Server().IsZero())
}

func TestParseShouldAcceptConfigWithoutServer(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	assert.True(t, config.Server().IsZero())
}

func TestParseShouldAcceptConfigWithoutQualityGate(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	qg := config.Projects()[0].Gate()
	assert.Empty(t, qg.Rules())
}

func TestParseShouldRejectEmptyName(t *testing.T) {
	data := []byte(`
name: ""
projects:
  - name: backend
    pattern: //backend/...
`)

	_, err := gavelconfig.Parse(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestParseShouldRejectMissingProjects(t *testing.T) {
	data := []byte(`
name: my-monorepo
`)

	_, err := gavelconfig.Parse(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one project")
}

func TestParseShouldRejectProjectWithEmptyName(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: ""
    pattern: //backend/...
`)

	_, err := gavelconfig.Parse(data)
	assert.Error(t, err)
}

func TestParseShouldRejectProjectWithEmptyPattern(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: ""
`)

	_, err := gavelconfig.Parse(data)
	assert.Error(t, err)
}

func TestParseShouldRejectDuplicatePatterns(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
  - name: backend-copy
    pattern: //backend/...
`)

	_, err := gavelconfig.Parse(data)
	assert.Error(t, err)
}

func TestParseShouldProduceQualityGateWithMinNewCodeCoverage(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      new_code_coverage:
        min: 80
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	qg := config.Projects()[0].Gate()
	rules := qg.Rules()
	require.Len(t, rules, 1)

	rule := rules[0]
	assert.Equal(t, evidence.SubtypeNewCodeCoverage, rule.Subtype())
	assert.IsType(t, qualitygate.MinNewCodeCoverage{}, rule.Strategy())
}


func TestParseShouldRejectInvalidMinNewCodeCoverage(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      new_code_coverage:
        min: 150
`)

	_, err := gavelconfig.Parse(data)
	assert.Error(t, err)
}

func TestParseShouldRejectInvalidMinCoverage(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      coverage:
        min: 150
`)

	_, err := gavelconfig.Parse(data)
	assert.Error(t, err)
}

func TestParseShouldRejectNegativeMaxFindings(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      findings:
        max_error: -1
`)

	_, err := gavelconfig.Parse(data)
	assert.Error(t, err)
}

func TestParseShouldLinkProjectRefsToGavelspace(t *testing.T) {
	data := []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
  - name: frontend
    pattern: //web/...
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	gs := config.Gavelspace()
	refs := gs.Projects()
	require.Len(t, refs, 2)

	assert.Equal(t, config.Projects()[0].ID(), refs[0].ID())
	assert.Equal(t, "//backend/...", refs[0].TargetPattern())

	assert.Equal(t, config.Projects()[1].ID(), refs[1].ID())
	assert.Equal(t, "//web/...", refs[1].TargetPattern())
}

func TestParseShouldAcceptFindingsSourceAuto(t *testing.T) {
	data := []byte(`
name: mono
findings_source: auto
projects:
  - name: svc
    pattern: //svc/...
`)

	config, err := gavelconfig.Parse(data)

	require.NoError(t, err)
	assert.Equal(t, "auto", config.FindingsSource())
}

func TestParseShouldAcceptFindingsSourceGavel(t *testing.T) {
	data := []byte(`
name: mono
findings_source: gavel
projects:
  - name: svc
    pattern: //svc/...
`)

	config, err := gavelconfig.Parse(data)

	require.NoError(t, err)
	assert.Equal(t, "gavel", config.FindingsSource())
}

func TestParseShouldAcceptFindingsSourceRulesLint(t *testing.T) {
	data := []byte(`
name: mono
findings_source: rules_lint
projects:
  - name: svc
    pattern: //svc/...
`)

	config, err := gavelconfig.Parse(data)

	require.NoError(t, err)
	assert.Equal(t, "rules_lint", config.FindingsSource())
}

func TestParseShouldDefaultFindingsSourceToEmpty(t *testing.T) {
	data := []byte(`
name: mono
projects:
  - name: svc
    pattern: //svc/...
`)

	config, err := gavelconfig.Parse(data)

	require.NoError(t, err)
	assert.Equal(t, "", config.FindingsSource())
}

func TestParseShouldAcceptCoverageOptions(t *testing.T) {
	data := []byte(`
name: mono
projects:
  - name: backend
    pattern: //backend/...
    coverage_options:
      test_size_filters: "small,medium"
      test_tag_filters: "-integration,-manual"
      instrumentation_filter: "//backend/..."
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	opts := config.CoverageOptionsForProject("backend")
	assert.Equal(t, "small,medium", opts.TestSizeFilters())
	assert.Equal(t, "-integration,-manual", opts.TestTagFilters())
	assert.Equal(t, "//backend/...", opts.InstrumentationFilter())
	assert.False(t, opts.IsZero())
}

func TestParseShouldAcceptEmptyCoverageOptions(t *testing.T) {
	data := []byte(`
name: mono
projects:
  - name: backend
    pattern: //backend/...
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	opts := config.CoverageOptionsForProject("backend")
	assert.True(t, opts.IsZero())
}

func TestParseShouldAcceptPartialCoverageOptions(t *testing.T) {
	data := []byte(`
name: mono
projects:
  - name: backend
    pattern: //backend/...
    coverage_options:
      test_size_filters: "small"
`)

	config, err := gavelconfig.Parse(data)
	require.NoError(t, err)

	opts := config.CoverageOptionsForProject("backend")
	assert.Equal(t, "small", opts.TestSizeFilters())
	assert.Equal(t, "", opts.TestTagFilters())
	assert.Equal(t, "", opts.InstrumentationFilter())
	assert.False(t, opts.IsZero())
}

func TestParseShouldRejectInvalidFindingsSource(t *testing.T) {
	data := []byte(`
name: mono
findings_source: unknown_source
projects:
  - name: svc
    pattern: //svc/...
`)

	_, err := gavelconfig.Parse(data)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "findings_source")
	assert.Contains(t, err.Error(), "unknown_source")
}

func TestParseFileNonexistentPathReturnsReadConfigError(t *testing.T) {
	_, err := gavelconfig.ParseFile("/nonexistent/path/gavel.yaml")
	require.Error(t, err)
	assert.ErrorIs(t, err, gavelconfig.ErrReadConfig)
	assert.Equal(t, failure.Validation, failure.Of(err))
}

func TestParseInvalidYAMLReturnsParseConfigError(t *testing.T) {
	_, err := gavelconfig.Parse([]byte(`{invalid yaml`))
	require.Error(t, err)
	assert.ErrorIs(t, err, gavelconfig.ErrParseConfig)
	assert.Equal(t, failure.Validation, failure.Of(err))
}

func TestParseDomainErrorsAreClassifiedAsParseConfig(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{
			name: "invalid gavelspace name",
			data: []byte(`
name: ""
projects:
  - name: backend
    pattern: //backend/...
`),
		},
		{
			name: "invalid project pattern",
			data: []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: ""
`),
		},
		{
			name: "invalid tooling language (empty)",
			data: []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    tooling:
      - "  "
`),
		},
		{
			name: "invalid quality_gate license (empty forbidden list)",
			data: []byte(`
name: my-monorepo
projects:
  - name: backend
    pattern: //backend/...
    quality_gate:
      license:
        forbidden: []
`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := gavelconfig.Parse(tc.data)
			require.Error(t, err)
			assert.ErrorIs(t, err, gavelconfig.ErrParseConfig,
				"every error path out of Parse must wrap ErrParseConfig")
			assert.Equal(t, failure.Validation, failure.Of(err))
		})
	}
}
