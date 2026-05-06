package qualitygate_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

func TestCountBySeverityEvaluatePassesForNonFindingContent(t *testing.T) {
	strategy, err := qualitygate.NewCountBySeverity(0, 0, 0)
	require.NoError(t, err)
	content, err := coverage.NewContent(100, 80, nil)
	require.NoError(t, err)

	outcome := strategy.Evaluate(content)
	assert.True(t, outcome.Passed())
}

func TestNewCountBySeverity(t *testing.T) {
	tests := []struct {
		name       string
		maxError   int
		maxWarning int
		maxNote    int
		wantErr    bool
	}{
		{name: "allZero", maxError: 0, maxWarning: 0, maxNote: 0},
		{name: "positiveValues", maxError: 5, maxWarning: 10, maxNote: 20},
		{name: "negativeError", maxError: -1, maxWarning: 0, maxNote: 0, wantErr: true},
		{name: "negativeWarning", maxError: 0, maxWarning: -1, maxNote: 0, wantErr: true},
		{name: "negativeNote", maxError: 0, maxWarning: 0, maxNote: -1, wantErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			_, err := qualitygate.NewCountBySeverity(tcase.maxError, tcase.maxWarning, tcase.maxNote)
			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, qualitygate.ErrInvalidStrategy)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestCountBySeverityEvaluate(t *testing.T) {
	tests := []struct {
		name       string
		maxError   int
		maxWarning int
		maxNote    int
		findings   []finding.Finding
		wantPass   bool
		wantDetail string
	}{
		{
			name:     "emptyFindings",
			findings: nil,
			wantPass: true,
		},
		{
			name:       "belowAllThresholds",
			maxError:   5,
			maxWarning: 5,
			maxNote:    5,
			findings:   buildFindings(t, 1, finding.SeverityError),
			wantPass:   true,
		},
		{
			name:       "atErrorThreshold",
			maxError:   2,
			maxWarning: 10,
			maxNote:    10,
			findings:   buildFindings(t, 2, finding.SeverityError),
			wantPass:   true,
		},
		{
			name:       "aboveErrorThreshold",
			maxError:   0,
			maxWarning: 10,
			maxNote:    10,
			findings:   buildFindings(t, 3, finding.SeverityError),
			wantPass:   false,
			wantDetail: "3 errors (max 0)",
		},
		{
			name:       "aboveWarningThreshold",
			maxError:   10,
			maxWarning: 1,
			maxNote:    10,
			findings:   buildFindings(t, 4, finding.SeverityWarning),
			wantPass:   false,
			wantDetail: "4 warnings (max 1)",
		},
		{
			name:       "aboveNoteThreshold",
			maxError:   10,
			maxWarning: 10,
			maxNote:    0,
			findings:   buildFindings(t, 2, finding.SeverityNote),
			wantPass:   false,
			wantDetail: "2 notes (max 0)",
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			strategy, err := qualitygate.NewCountBySeverity(tcase.maxError, tcase.maxWarning, tcase.maxNote)
			require.NoError(t, err)

			content := buildFindingsContent(t, tcase.findings)
			outcome := strategy.Evaluate(content)

			assert.Equal(t, tcase.wantPass, outcome.Passed())
			if tcase.wantDetail != "" {
				assert.Equal(t, tcase.wantDetail, outcome.Detail())
			}
		})
	}
}

func buildFindings(t *testing.T, count int, severity finding.Severity) []finding.Finding {
	t.Helper()
	findings := make([]finding.Finding, 0, count)
	for i := range count {
		fp, err := finding.NewFingerprintID(fmt.Sprintf("fp%d", i))
		require.NoError(t, err)
		f, err := finding.NewFinding("tool", "rule", severity, "f.go", 1, "msg", fp)
		require.NoError(t, err)
		findings = append(findings, f)
	}
	return findings
}

func buildFindingsContent(t *testing.T, findings []finding.Finding) finding.Content {
	t.Helper()
	findCont, err := finding.NewContent(evidence.SubtypeCodeQuality, findings)
	require.NoError(t, err)
	return findCont
}
