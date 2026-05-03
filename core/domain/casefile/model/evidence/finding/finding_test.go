package finding_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
)

func TestNewFinding(t *testing.T) {
	validSeverity := finding.SeverityError
	validFingerprint, err := finding.NewFingerprintID("abc123")
	require.NoError(t, err)

	tests := []struct {
		name        string
		tool        string
		ruleID      string
		severity    finding.Severity
		filePath    string
		line        int
		message     string
		fingerprint finding.FingerprintID
		wantErr     bool
	}{
		{
			name:        "shouldCreateValidFinding",
			tool:        "pmd",
			ruleID:      "UnusedVariable",
			severity:    validSeverity,
			filePath:    "src/Main.java",
			line:        42,
			message:     "unused variable x",
			fingerprint: validFingerprint,
		},
		{
			name:        "shouldCreateValidFindingWithEmptyMessage",
			tool:        "pmd",
			ruleID:      "UnusedVariable",
			severity:    validSeverity,
			filePath:    "src/Main.java",
			line:        1,
			message:     "",
			fingerprint: validFingerprint,
		},
		{
			name:        "shouldCreateValidFindingWithZeroLine",
			tool:        "pmd",
			ruleID:      "UnusedVariable",
			severity:    validSeverity,
			filePath:    "src/Main.java",
			line:        0,
			message:     "some message",
			fingerprint: validFingerprint,
		},
		{
			name:        "shouldRejectEmptyTool",
			tool:        "",
			ruleID:      "UnusedVariable",
			severity:    validSeverity,
			filePath:    "src/Main.java",
			line:        1,
			message:     "msg",
			fingerprint: validFingerprint,
			wantErr:     true,
		},
		{
			name:        "shouldRejectBlankTool",
			tool:        "   ",
			ruleID:      "UnusedVariable",
			severity:    validSeverity,
			filePath:    "src/Main.java",
			line:        1,
			message:     "msg",
			fingerprint: validFingerprint,
			wantErr:     true,
		},
		{
			name:        "shouldRejectEmptyRuleID",
			tool:        "pmd",
			ruleID:      "",
			severity:    validSeverity,
			filePath:    "src/Main.java",
			line:        1,
			message:     "msg",
			fingerprint: validFingerprint,
			wantErr:     true,
		},
		{
			name:        "shouldRejectEmptyFilePath",
			tool:        "pmd",
			ruleID:      "UnusedVariable",
			severity:    validSeverity,
			filePath:    "",
			line:        1,
			message:     "msg",
			fingerprint: validFingerprint,
			wantErr:     true,
		},
		{
			name:        "shouldRejectNegativeLine",
			tool:        "pmd",
			ruleID:      "UnusedVariable",
			severity:    validSeverity,
			filePath:    "src/Main.java",
			line:        -1,
			message:     "msg",
			fingerprint: validFingerprint,
			wantErr:     true,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			fnd, err := finding.NewFinding(tcase.tool, tcase.ruleID, tcase.severity, tcase.filePath, tcase.line, tcase.message, tcase.fingerprint)

			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, finding.ErrInvalidFinding)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tcase.tool, fnd.Tool())
			assert.Equal(t, tcase.ruleID, fnd.RuleID())
			assert.Equal(t, tcase.severity, fnd.Severity())
			assert.Equal(t, tcase.filePath, fnd.FilePath())
			assert.Equal(t, tcase.line, fnd.Line())
			assert.Equal(t, tcase.message, fnd.Message())
			assert.Equal(t, tcase.fingerprint, fnd.ID())
		})
	}
}

func TestFindingEqual(t *testing.T) {
	fp1, err := finding.NewFingerprintID("abc123")
	require.NoError(t, err)
	fp2, err := finding.NewFingerprintID("abc123")
	require.NoError(t, err)
	fpDifferent, err := finding.NewFingerprintID("xyz789")
	require.NoError(t, err)

	find1, err := finding.NewFinding("pmd", "Rule1", finding.SeverityError, "a.java", 1, "msg", fp1)
	require.NoError(t, err)
	find2, err := finding.NewFinding("spotbugs", "Rule2", finding.SeverityWarning, "b.java", 99, "other", fp2)
	require.NoError(t, err)
	f3, err := finding.NewFinding("pmd", "Rule1", finding.SeverityError, "a.java", 1, "msg", fpDifferent)
	require.NoError(t, err)

	assert.True(t, find1.Equal(find2))
	assert.False(t, find1.Equal(f3))
}
