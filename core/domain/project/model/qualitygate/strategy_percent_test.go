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

func TestMinPercentageEvaluatePassesForNonCoverageContent(t *testing.T) {
	strategy, err := qualitygate.NewMinPercentage(80)
	require.NoError(t, err)
	content, err := finding.NewContent(evidence.SubtypeCodeQuality, nil)
	require.NoError(t, err)

	outcome := strategy.Evaluate(content)
	assert.True(t, outcome.Passed())
}

func TestNewMinPercentage(t *testing.T) {
	tests := []struct {
		name    string
		min     float64
		wantErr bool
	}{
		{name: "validZero", min: 0},
		{name: "validHundred", min: 100},
		{name: "validMiddle", min: 75.5},
		{name: "negativeValue", min: -1, wantErr: true},
		{name: "aboveHundred", min: 100.1, wantErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			_, err := qualitygate.NewMinPercentage(tcase.min)
			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, qualitygate.ErrInvalidStrategy)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMinPercentageEvaluate(t *testing.T) {
	tests := []struct {
		name         string
		min          float64
		totalLines   int
		coveredLines int
		wantPass     bool
		wantDetail   string
	}{
		{
			name:         "aboveMin",
			min:          80,
			totalLines:   100,
			coveredLines: 90,
			wantPass:     true,
		},
		{
			name:         "exactlyAtMin",
			min:          80,
			totalLines:   100,
			coveredLines: 80,
			wantPass:     true,
		},
		{
			name:         "belowMin",
			min:          80,
			totalLines:   100,
			coveredLines: 75,
			wantPass:     false,
			wantDetail:   "75.0% coverage (min 80.0%)",
		},
		{
			name:         "zeroCoverageWithPositiveMin",
			min:          50,
			totalLines:   100,
			coveredLines: 0,
			wantPass:     false,
			wantDetail:   "0.0% coverage (min 50.0%)",
		},
		{
			name:         "zeroCoverageWithZeroMin",
			min:          0,
			totalLines:   100,
			coveredLines: 0,
			wantPass:     true,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			strategy, err := qualitygate.NewMinPercentage(tcase.min)
			require.NoError(t, err)

			content, err := coverage.NewContent(tcase.totalLines, tcase.coveredLines, nil)
			require.NoError(t, err)

			outcome := strategy.Evaluate(content)
			assert.Equal(t, tcase.wantPass, outcome.Passed())
			if tcase.wantDetail != "" {
				assert.Equal(t, tcase.wantDetail, outcome.Detail())
			}
		})
	}
}
