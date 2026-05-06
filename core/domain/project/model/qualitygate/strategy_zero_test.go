package qualitygate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

func TestZeroToleranceEvaluatePassesForNonFindingContent(t *testing.T) {
	strategy := qualitygate.NewZeroTolerance()
	content, err := coverage.NewContent(100, 80, nil)
	require.NoError(t, err)

	outcome := strategy.Evaluate(content)
	assert.True(t, outcome.Passed())
}

func TestZeroToleranceEvaluate(t *testing.T) {
	tests := []struct {
		name     string
		findings []finding.Finding
		wantPass bool
	}{
		{
			name:     "zeroFindings",
			findings: nil,
			wantPass: true,
		},
		{
			name:     "oneFinding",
			findings: buildFindings(t, 1, finding.SeverityError),
			wantPass: false,
		},
		{
			name:     "multipleFindings",
			findings: buildFindings(t, 5, finding.SeverityWarning),
			wantPass: false,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			strategy := qualitygate.NewZeroTolerance()
			content, err := finding.NewContent(evidence.SubtypeCodeQuality, tcase.findings)
			require.NoError(t, err)

			outcome := strategy.Evaluate(content)
			assert.Equal(t, tcase.wantPass, outcome.Passed())
		})
	}
}
