package model_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

var (
	testTime   = time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	testTenant = tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))
)

func TestUpdateExcludePatterns_StoresValidPatternsWithinScope(t *testing.T) {
	p, err := model.NewProject(testTenant, "core", "core", "//core/...")
	require.NoError(t, err)

	require.NoError(t, p.UpdateExcludePatterns([]string{"//core/gen/..."}, testTime))

	assert.Equal(t, []string{"//core/gen/..."}, p.ExcludePatterns())
}

func TestUpdateExcludePatterns_RejectsPatternOutsideScope(t *testing.T) {
	p, err := model.NewProject(testTenant, "core", "core", "//core/...")
	require.NoError(t, err)

	err = p.UpdateExcludePatterns([]string{"//apps/cli/..."}, testTime)

	require.Error(t, err)
}

func TestUpdateExcludePatterns_RejectsMalformedPattern(t *testing.T) {
	p, err := model.NewProject(testTenant, "core", "core", "//core/...")
	require.NoError(t, err)

	err = p.UpdateExcludePatterns([]string{"not a pattern"}, testTime)

	require.Error(t, err)
}

func TestExcludePatterns_ReturnsDefensiveCopy(t *testing.T) {
	project, err := model.NewProject(testTenant, "core", "core", "//core/...")
	require.NoError(t, err)
	require.NoError(t, project.UpdateExcludePatterns([]string{"//core/gen/..."}, testTime))

	got := project.ExcludePatterns()
	got[0] = "tampered"

	assert.Equal(t, []string{"//core/gen/..."}, project.ExcludePatterns())
}

func TestNewProject(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		projectName   string
		targetPattern string
		expectErr     bool
	}{
		{
			name:          "valid project",
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "//my-service/...",
			expectErr:     false,
		},
		{
			name:          "empty target pattern rejected",
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "",
			expectErr:     true,
		},
		{
			name:          "whitespace target pattern rejected",
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "   ",
			expectErr:     true,
		},
		{
			name:          "target pattern without double slash rejected",
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "foo/...",
			expectErr:     true,
		},
		{
			name:          "bare double slash rejected",
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "//",
			expectErr:     true,
		},
		{
			name:          "target pattern with spaces rejected",
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "// foo/...",
			expectErr:     true,
		},
		{
			name:          "valid recursive pattern",
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "//foo/bar/...",
			expectErr:     false,
		},
		{
			name:          "valid workspace-wide pattern",
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "//...",
			expectErr:     false,
		},
		{
			name:          "valid explicit target",
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "//foo/bar:mylib",
			expectErr:     false,
		},
		{
			name:          "valid all target",
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "//foo/bar:all",
			expectErr:     false,
		},
		{
			name:          "valid shorthand package",
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "//foo/bar",
			expectErr:     false,
		},
		{
			name:          "valid pattern with underscore and dot",
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "//third_party/com.example/...",
			expectErr:     false,
		},
		{
			name:          "empty name rejected",
			key:           "my-service",
			projectName:   "",
			targetPattern: "//my-service/...",
			expectErr:     true,
		},
		{
			name:          "whitespace name rejected",
			key:           "my-service",
			projectName:   "   ",
			targetPattern: "//my-service/...",
			expectErr:     true,
		},
		{
			name:          "empty key rejected",
			key:           "",
			projectName:   "my-service",
			targetPattern: "//my-service/...",
			expectErr:     true,
		},
		{
			name:          "uppercase key rejected",
			key:           "My-Service",
			projectName:   "my-service",
			targetPattern: "//my-service/...",
			expectErr:     true,
		},
		{
			name:          "key with spaces rejected",
			key:           "my service",
			projectName:   "my-service",
			targetPattern: "//my-service/...",
			expectErr:     true,
		},
		{
			name:          "key starting with hyphen rejected",
			key:           "-my-service",
			projectName:   "my-service",
			targetPattern: "//my-service/...",
			expectErr:     true,
		},
		{
			name:          "key ending with hyphen rejected",
			key:           "my-service-",
			projectName:   "my-service",
			targetPattern: "//my-service/...",
			expectErr:     true,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			project, err := model.NewProject(testTenant, tcase.key, tcase.projectName, tcase.targetPattern)

			if tcase.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, model.ErrInvalidProject)
				return
			}

			require.NoError(t, err)
			assert.True(t, testTenant.Equal(project.TenantID()))
			assert.Equal(t, tcase.key, project.Key())
			assert.Equal(t, tcase.projectName, project.Name())
			assert.Equal(t, tcase.targetPattern, project.TargetPattern())
			assert.Equal(t, "main", project.DefaultBranch())
		})
	}
}

func TestNewProjectGeneratesUniqueIDs(t *testing.T) {
	proj1, err := model.NewProject(testTenant, "service-a", "service-a", "//service-a/...")
	require.NoError(t, err)

	p2, err := model.NewProject(testTenant, "service-b", "service-b", "//service-b/...")
	require.NoError(t, err)

	assert.False(t, proj1.ID().Equal(p2.ID()))
}

func TestProjectUpdateQualityGate(t *testing.T) {
	project, err := model.NewProject(testTenant, "my-service", "my-service", "//my-service/...")
	require.NoError(t, err)
	project.ClearEvents()

	qGate, err := qualitygate.NewGate(nil)
	require.NoError(t, err)

	project.UpdateQualityGate(qGate, testTime)

	assert.Equal(t, qGate, project.Gate())

	events := project.Events()
	require.Len(t, events, 1)
	evt, ok := events[0].(model.QualityGateUpdated)
	require.True(t, ok)
	assert.True(t, project.ID().Equal(evt.ProjectID()))
}

func TestProjectUpdateLanguages(t *testing.T) {
	project, err := model.NewProject(testTenant, "my-service", "my-service", "//my-service/...")
	require.NoError(t, err)
	project.ClearEvents()

	java, err := coverage.NewLanguage("java")
	require.NoError(t, err)
	golang, err := coverage.NewLanguage("go")
	require.NoError(t, err)

	languages := []coverage.Language{java, golang}
	project.UpdateLanguages(languages, testTime)

	got := project.Languages()
	assert.Equal(t, languages, got)

	events := project.Events()
	require.Len(t, events, 1)
	evt, ok := events[0].(model.LanguagesUpdated)
	require.True(t, ok)
	assert.True(t, project.ID().Equal(evt.ProjectID()))
}

func TestProjectUpdateToolSelection(t *testing.T) {
	project, err := model.NewProject(testTenant, "my-service", "my-service", "//my-service/...")
	require.NoError(t, err)
	project.ClearEvents()

	selection := map[string][]string{"go": {"golangci-lint", "archtest"}}
	project.UpdateToolSelection(selection, testTime)

	assert.Equal(t, selection, project.ToolSelection())

	events := project.Events()
	require.Len(t, events, 1)
	evt, ok := events[0].(model.ToolSelectionUpdated)
	require.True(t, ok)
	assert.True(t, project.ID().Equal(evt.ProjectID()))
}

func TestProjectToolSelectionReturnsDefensiveCopy(t *testing.T) {
	project, err := model.NewProject(testTenant, "my-service", "my-service", "//my-service/...")
	require.NoError(t, err)
	project.UpdateToolSelection(map[string][]string{"go": {"golangci-lint"}}, testTime)

	mutated := project.ToolSelection()
	mutated["go"][0] = "tampered"
	mutated["java"] = []string{"pmd"}

	assert.Equal(t, map[string][]string{"go": {"golangci-lint"}}, project.ToolSelection())
}

func TestProjectUpdateTargetPattern(t *testing.T) {
	project, err := model.NewProject(testTenant, "my-service", "my-service", "//my-service/...")
	require.NoError(t, err)
	project.ClearEvents()

	require.NoError(t, project.UpdateTargetPattern("//my-service/viol2/...", testTime))

	assert.Equal(t, "//my-service/viol2/...", project.TargetPattern())

	events := project.Events()
	require.Len(t, events, 1)
	evt, ok := events[0].(model.TargetPatternUpdated)
	require.True(t, ok)
	assert.True(t, project.ID().Equal(evt.ProjectID()))
	assert.Equal(t, testTime, evt.OccurredAt())
	assert.Equal(t, model.EventNameTargetPatternUpdated, evt.EventName())
}

func TestProjectUpdateTargetPatternRejectsInvalid(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
	}{
		{name: "empty", pattern: ""},
		{name: "whitespace", pattern: "   "},
		{name: "missing leading slashes", pattern: "foo/..."},
		{name: "bare double slash", pattern: "//"},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			project, err := model.NewProject(testTenant, "my-service", "my-service", "//my-service/...")
			require.NoError(t, err)
			project.ClearEvents()

			err = project.UpdateTargetPattern(tcase.pattern, testTime)
			require.Error(t, err)
			assert.ErrorIs(t, err, model.ErrInvalidProject)

			assert.Equal(t, "//my-service/...", project.TargetPattern(),
				"invalid update leaves prior pattern intact")
			assert.Empty(t, project.Events(), "no event recorded on invalid update")
		})
	}
}

func TestProjectLanguagesDefensiveCopy(t *testing.T) {
	project, err := model.NewProject(testTenant, "my-service", "my-service", "//my-service/...")
	require.NoError(t, err)

	java, err := coverage.NewLanguage("java")
	require.NoError(t, err)
	golang, err := coverage.NewLanguage("go")
	require.NoError(t, err)

	input := []coverage.Language{java, golang}
	project.UpdateLanguages(input, testTime)

	returned := project.Languages()
	returned[0] = golang

	assert.Equal(t, java, project.Languages()[0])
}

func TestReconstituteProject(t *testing.T) {
	java, err := coverage.NewLanguage("java")
	require.NoError(t, err)

	qGate, err := qualitygate.NewGate(nil)
	require.NoError(t, err)
	validID := model.NewProjectID(uuid.New())

	tests := []struct {
		name          string
		id            model.ProjectID
		key           string
		projectName   string
		targetPattern string
		defaultBranch string
		languages     []coverage.Language
		qualityGate   qualitygate.Gate
		expectErr     bool
	}{
		{
			name:          "valid reconstitution",
			id:            validID,
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "//my-service/...",
			defaultBranch: "develop",
			languages:     []coverage.Language{java},
			qualityGate:   qGate,
			expectErr:     false,
		},
		{
			name:          "empty target pattern rejected",
			id:            validID,
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "",
			defaultBranch: "main",
			languages:     []coverage.Language{},
			qualityGate:   qualitygate.Gate{},
			expectErr:     true,
		},
		{
			name:          "invalid target pattern rejected",
			id:            validID,
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "foo/bar",
			defaultBranch: "main",
			languages:     []coverage.Language{},
			qualityGate:   qualitygate.Gate{},
			expectErr:     true,
		},
		{
			name:          "empty key rejected",
			id:            validID,
			key:           "",
			projectName:   "my-service",
			targetPattern: "//my-service/...",
			defaultBranch: "main",
			languages:     nil,
			qualityGate:   qualitygate.Gate{},
			expectErr:     true,
		},
		{
			name:          "empty name rejected",
			id:            validID,
			key:           "my-service",
			projectName:   "",
			targetPattern: "//my-service/...",
			defaultBranch: "main",
			languages:     nil,
			qualityGate:   qualitygate.Gate{},
			expectErr:     true,
		},
		{
			name:          "empty branch rejected",
			id:            validID,
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "//my-service/...",
			defaultBranch: "",
			languages:     nil,
			qualityGate:   qualitygate.Gate{},
			expectErr:     true,
		},
		{
			name:          "blank branch rejected",
			id:            validID,
			key:           "my-service",
			projectName:   "my-service",
			targetPattern: "//my-service/...",
			defaultBranch: "   ",
			languages:     nil,
			qualityGate:   qualitygate.Gate{},
			expectErr:     true,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			project, err := model.ReconstituteProject(
				tcase.id, testTenant, tcase.key, tcase.projectName, tcase.targetPattern,
				tcase.defaultBranch, tcase.languages, tcase.qualityGate, nil, nil,
			)

			if tcase.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, model.ErrInvalidProject)
				return
			}

			require.NoError(t, err)
			assert.True(t, tcase.id.Equal(project.ID()))
			assert.True(t, testTenant.Equal(project.TenantID()))
			assert.Equal(t, tcase.key, project.Key())
			assert.Equal(t, tcase.projectName, project.Name())
			assert.Equal(t, tcase.targetPattern, project.TargetPattern())
			assert.Equal(t, tcase.defaultBranch, project.DefaultBranch())
			assert.Equal(t, tcase.languages, project.Languages())
			assert.Equal(t, tcase.qualityGate, project.Gate())
		})
	}
}

func TestReconstituteProjectWithBaselines(t *testing.T) {
	projectID := model.NewProjectID(uuid.New())
	qGate, err := qualitygate.NewGate(nil)
	require.NoError(t, err)

	cov := 85.5
	baselines := map[string]model.Baseline{
		"main":    model.NewBaseline([]string{"fp-1", "fp-2"}, []string{"arch-1"}, &cov, nil),
		"develop": model.NewBaseline([]string{"fp-3"}, nil, nil, nil),
	}

	project, err := model.ReconstituteProject(
		projectID, testTenant, "svc", "svc", "//svc/...", "main",
		nil, qGate, nil, baselines,
	)
	require.NoError(t, err)

	mainBL := project.Baseline("main")
	assert.Equal(t, []string{"fp-1", "fp-2"}, mainBL.Fingerprints())
	assert.Equal(t, []string{"arch-1"}, mainBL.ArchIDs())
	require.NotNil(t, mainBL.CoveragePercent())
	assert.InDelta(t, 85.5, *mainBL.CoveragePercent(), 0.01)
	assert.True(t, mainBL.HasPrevious())

	devBL := project.Baseline("develop")
	assert.Equal(t, []string{"fp-3"}, devBL.Fingerprints())
	assert.True(t, devBL.HasPrevious())

	featureBL := project.Baseline("nonexistent")
	assert.True(t, featureBL.HasPrevious(),
		"unknown branch falls back to default-branch baseline")
	assert.Equal(t, []string{"fp-1", "fp-2"}, featureBL.Fingerprints())
}

func TestBaselineFallbackToDefaultBranch(t *testing.T) {
	projectID := model.NewProjectID(uuid.New())
	qGate, err := qualitygate.NewGate(nil)
	require.NoError(t, err)

	cov := 72.3
	baselines := map[string]model.Baseline{
		"main": model.NewBaseline([]string{"fp-1", "fp-2"}, []string{"arch-1"}, &cov, nil),
	}

	project, err := model.ReconstituteProject(
		projectID, testTenant, "svc", "svc", "//svc/...", "main",
		nil, qGate, nil, baselines,
	)
	require.NoError(t, err)

	t.Run("shouldReturnBaselineForExactBranch", func(t *testing.T) {
		bl := project.Baseline("main")
		assert.True(t, bl.HasPrevious())
		assert.Equal(t, []string{"fp-1", "fp-2"}, bl.Fingerprints())
	})

	t.Run("shouldFallbackToDefaultBranchWhenBranchNotFound", func(t *testing.T) {
		bl := project.Baseline("feature-x")
		assert.True(t, bl.HasPrevious())
		assert.Equal(t, []string{"fp-1", "fp-2"}, bl.Fingerprints())
		assert.Equal(t, []string{"arch-1"}, bl.ArchIDs())
		require.NotNil(t, bl.CoveragePercent())
		assert.InDelta(t, 72.3, *bl.CoveragePercent(), 0.001)
	})

	t.Run("shouldReturnEmptyWhenDefaultBranchHasNoBaseline", func(t *testing.T) {
		empty, err := model.NewProject(testTenant, "empty", "empty", "//empty/...")
		require.NoError(t, err)
		bl := empty.Baseline("main")
		assert.False(t, bl.HasPrevious())
	})

	t.Run("shouldPreferExactBranchOverFallback", func(t *testing.T) {
		p, err := model.NewProject(testTenant, "multi", "multi", "//multi/...")
		require.NoError(t, err)
		p.UpdateBaseline("main", []string{"main-fp"}, nil, nil, nil)
		p.UpdateBaseline("feature-x", []string{"feature-fp"}, nil, nil, nil)

		bl := p.Baseline("feature-x")
		assert.Equal(t, []string{"feature-fp"}, bl.Fingerprints())
	})
}

func TestNewProjectRejectsZeroTenant(t *testing.T) {
	_, err := model.NewProject(tenant.TenantID{}, "svc", "svc", "//svc/...")
	require.Error(t, err)
	assert.ErrorIs(t, err, model.ErrInvalidProject)
}
