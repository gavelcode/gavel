package qualitygate_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/license"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

func TestForbiddenListEvaluatePassesForNonLicenseContent(t *testing.T) {
	strategy, err := qualitygate.NewForbiddenList([]string{"GPL-3.0"})
	require.NoError(t, err)
	content, err := finding.NewContent(evidence.SubtypeCodeQuality, nil)
	require.NoError(t, err)

	outcome := strategy.Evaluate(content)
	assert.True(t, outcome.Passed())
}

func TestNewForbiddenList(t *testing.T) {
	tests := []struct {
		name      string
		forbidden []string
		wantErr   bool
	}{
		{name: "validSingle", forbidden: []string{"GPL-3.0"}},
		{name: "validMultiple", forbidden: []string{"GPL-3.0", "AGPL-3.0"}},
		{name: "emptyList", forbidden: []string{}, wantErr: true},
		{name: "nilList", forbidden: nil, wantErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			_, err := qualitygate.NewForbiddenList(tcase.forbidden)
			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, qualitygate.ErrInvalidStrategy)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestForbiddenListEvaluate(t *testing.T) {
	tests := []struct {
		name       string
		forbidden  []string
		deps       []license.Dependency
		wantPass   bool
		wantDetail string
	}{
		{
			name:      "noForbiddenMatch",
			forbidden: []string{"GPL-3.0"},
			deps:      buildDeps(t, "MIT", "Apache-2.0"),
			wantPass:  true,
		},
		{
			name:       "oneMatch",
			forbidden:  []string{"GPL-3.0"},
			deps:       buildDeps(t, "MIT", "GPL-3.0"),
			wantPass:   false,
			wantDetail: "forbidden licenses: dep-1 (GPL-3.0)",
		},
		{
			name:      "emptyDeps",
			forbidden: []string{"GPL-3.0"},
			deps:      nil,
			wantPass:  true,
		},
		{
			name:       "multipleMatches",
			forbidden:  []string{"GPL-3.0", "AGPL-3.0"},
			deps:       buildDeps(t, "GPL-3.0", "MIT", "AGPL-3.0"),
			wantPass:   false,
			wantDetail: "forbidden licenses: dep-0 (GPL-3.0), dep-2 (AGPL-3.0)",
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			strategy, err := qualitygate.NewForbiddenList(tcase.forbidden)
			require.NoError(t, err)

			content, err := license.NewContent(tcase.deps)
			require.NoError(t, err)

			outcome := strategy.Evaluate(content)
			assert.Equal(t, tcase.wantPass, outcome.Passed())
			if tcase.wantDetail != "" {
				assert.Equal(t, tcase.wantDetail, outcome.Detail())
			}
		})
	}
}

func buildDeps(t *testing.T, licenses ...string) []license.Dependency {
	t.Helper()
	deps := make([]license.Dependency, 0, len(licenses))
	for i, lic := range licenses {
		depLic, err := license.NewDependency(fmt.Sprintf("dep-%d", i), "1.0.0", lic)
		require.NoError(t, err)
		deps = append(deps, depLic)
	}
	return deps
}
