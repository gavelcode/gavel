package coverage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
)

func TestNewLanguageCoverage(t *testing.T) {
	validLang, err := coverage.NewLanguage("go")
	require.NoError(t, err)

	tests := []struct {
		name         string
		language     coverage.Language
		totalLines   int
		coveredLines int
		wantErr      bool
	}{
		{
			name:         "shouldCreateValidLanguageCoverage",
			language:     validLang,
			totalLines:   100,
			coveredLines: 80,
		},
		{
			name:         "shouldCreateValidWithZeroLines",
			language:     validLang,
			totalLines:   0,
			coveredLines: 0,
		},
		{
			name:         "shouldRejectNegativeTotalLines",
			language:     validLang,
			totalLines:   -1,
			coveredLines: 0,
			wantErr:      true,
		},
		{
			name:         "shouldRejectNegativeCoveredLines",
			language:     validLang,
			totalLines:   100,
			coveredLines: -1,
			wantErr:      true,
		},
		{
			name:         "shouldRejectCoveredLinesGreaterThanTotalLines",
			language:     validLang,
			totalLines:   50,
			coveredLines: 51,
			wantErr:      true,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			langCov, err := coverage.NewLanguageStats(tcase.language, tcase.totalLines, tcase.coveredLines)

			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, coverage.ErrInvalidLanguageStats)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tcase.language, langCov.Language())
			assert.Equal(t, tcase.totalLines, langCov.TotalLines())
			assert.Equal(t, tcase.coveredLines, langCov.CoveredLines())
		})
	}
}

func TestLanguageCoveragePercent(t *testing.T) {
	lang, err := coverage.NewLanguage("go")
	require.NoError(t, err)

	tests := []struct {
		name         string
		totalLines   int
		coveredLines int
		expected     float64
	}{
		{
			name:         "shouldCalculatePercentNormally",
			totalLines:   200,
			coveredLines: 150,
			expected:     75.0,
		},
		{
			name:         "shouldReturnZeroWhenTotalLinesIsZero",
			totalLines:   0,
			coveredLines: 0,
			expected:     0.0,
		},
		{
			name:         "shouldReturn100WhenFullyCovered",
			totalLines:   100,
			coveredLines: 100,
			expected:     100.0,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			langCov, err := coverage.NewLanguageStats(lang, tcase.totalLines, tcase.coveredLines)
			require.NoError(t, err)

			assert.InDelta(t, tcase.expected, langCov.Percent(), 0.001)
		})
	}
}
