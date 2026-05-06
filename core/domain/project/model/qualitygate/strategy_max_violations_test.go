package qualitygate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/architecture"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

func TestMaxViolationsEvaluate(t *testing.T) {
	tests := []struct {
		name       string
		max        int
		violations int
		wantPass   bool
	}{
		{
			name:       "zeroViolationsWithZeroMax",
			max:        0,
			violations: 0,
			wantPass:   true,
		},
		{
			name:       "oneViolationWithZeroMax",
			max:        0,
			violations: 1,
			wantPass:   false,
		},
		{
			name:       "violationsAtMax",
			max:        3,
			violations: 3,
			wantPass:   true,
		},
		{
			name:       "violationsExceedMax",
			max:        3,
			violations: 4,
			wantPass:   false,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			strategy, err := qualitygate.NewMaxViolations(tcase.max)
			require.NoError(t, err)

			content := buildArchContent(t, tcase.violations)

			outcome := strategy.Evaluate(content)
			assert.Equal(t, tcase.wantPass, outcome.Passed())
		})
	}
}

func TestMaxViolationsNonArchitectureContentPasses(t *testing.T) {
	strategy, err := qualitygate.NewMaxViolations(0)
	require.NoError(t, err)

	content, err := coverage.NewContent(100, 80, nil)
	require.NoError(t, err)

	outcome := strategy.Evaluate(content)
	assert.True(t, outcome.Passed())
}

func TestNewMaxViolationsRejectsNegative(t *testing.T) {
	_, err := qualitygate.NewMaxViolations(-1)
	require.Error(t, err)
}

func buildArchContent(t *testing.T, count int) architecture.Content {
	t.Helper()
	violations := make([]architecture.Violation, 0, count)
	for i := range count {
		archViol, err := architecture.NewViolation(
			"layer-dependency",
			"domain/pkg"+string(rune('A'+i)),
			"infra/pkg"+string(rune('A'+i)),
			"violation",
		)
		require.NoError(t, err)
		violations = append(violations, archViol)
	}
	archCont, err := architecture.NewContent(violations)
	require.NoError(t, err)
	return archCont
}
