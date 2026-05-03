package architecture_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/architecture"
)

func validArchViolation(t *testing.T, rule, source, target string) architecture.Violation {
	t.Helper()
	archViol, err := architecture.NewViolation(rule, source, target, source+" imports "+target)
	require.NoError(t, err)
	return archViol
}

func TestNewArchitectureContent(t *testing.T) {
	viol1 := validArchViolation(t, "layer-dependency", "domain/foo", "infra/bar")

	tests := []struct {
		name       string
		violations []architecture.Violation
	}{
		{
			name:       "shouldCreateWithViolations",
			violations: []architecture.Violation{viol1},
		},
		{
			name:       "shouldCreateWithEmptyViolations",
			violations: []architecture.Violation{},
		},
		{
			name:       "shouldCreateWithNilViolations",
			violations: nil,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			archCont, err := architecture.NewContent(tcase.violations)

			require.NoError(t, err)
			if tcase.violations == nil {
				assert.Empty(t, archCont.Violations())
			} else {
				assert.Equal(t, len(tcase.violations), len(archCont.Violations()))
			}
		})
	}
}

func TestArchitectureContentType(t *testing.T) {
	archCont, err := architecture.NewContent(nil)
	require.NoError(t, err)

	assert.Equal(t, evidence.TypeArchitecture, archCont.Type())
}

func TestArchitectureContentSubtype(t *testing.T) {
	archCont, err := architecture.NewContent(nil)
	require.NoError(t, err)

	assert.Equal(t, evidence.SubtypeArchitecture, archCont.Subtype())
}

func TestArchitectureContentDefensiveCopy(t *testing.T) {
	viol1 := validArchViolation(t, "layer-dependency", "domain/foo", "infra/bar")
	viol2 := validArchViolation(t, "circular-dependency", "domain/a", "domain/b")
	original := []architecture.Violation{viol1}

	archCont, err := architecture.NewContent(original)
	require.NoError(t, err)

	original[0] = viol2
	assert.NotEqual(t, viol2, archCont.Violations()[0])

	returned := archCont.Violations()
	returned[0] = viol2
	assert.NotEqual(t, viol2, archCont.Violations()[0])
}

func TestArchitectureContentMerge(t *testing.T) {
	viol1 := validArchViolation(t, "layer-dependency", "domain/foo", "infra/bar")
	viol2 := validArchViolation(t, "layer-dependency", "domain/baz", "infra/qux")
	v3 := validArchViolation(t, "circular-dependency", "app/a", "app/b")

	ac1, err := architecture.NewContent([]architecture.Violation{viol1, viol2})
	require.NoError(t, err)
	ac2, err := architecture.NewContent([]architecture.Violation{v3})
	require.NoError(t, err)

	merged, err := ac1.Merge(ac2)
	require.NoError(t, err)
	mergedAC, ok := merged.(architecture.Content)
	require.True(t, ok)

	assert.Equal(t, 3, len(mergedAC.Violations()))
}
