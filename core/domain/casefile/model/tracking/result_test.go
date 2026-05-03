package tracking_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/tracking"
)

func TestClassifyFindings(t *testing.T) {
	fpA := mustFingerprintID(t, "fp-a")
	fpB := mustFingerprintID(t, "fp-b")
	fpC := mustFingerprintID(t, "fp-c")

	findingA := mustFinding(t, "tool", "rule1", finding.SeverityError, "file.go", 1, "msg-a", fpA)
	findingB := mustFinding(t, "tool", "rule2", finding.SeverityWarning, "file.go", 2, "msg-b", fpB)
	findingC := mustFinding(t, "tool", "rule3", finding.SeverityNote, "file.go", 3, "msg-c", fpC)

	tests := []struct {
		name                 string
		current              []finding.Finding
		previousFingerprints []finding.FingerprintID
		wantNewCount         int
		wantExistingCount    int
		wantResolvedCount    int
	}{
		{
			name:                 "all new when no previous fingerprints",
			current:              []finding.Finding{findingA, findingB},
			previousFingerprints: nil,
			wantNewCount:         2,
			wantExistingCount:    0,
			wantResolvedCount:    0,
		},
		{
			name:                 "all existing when full overlap",
			current:              []finding.Finding{findingA, findingB},
			previousFingerprints: []finding.FingerprintID{fpA, fpB},
			wantNewCount:         0,
			wantExistingCount:    2,
			wantResolvedCount:    0,
		},
		{
			name:                 "mixed new and existing",
			current:              []finding.Finding{findingA, findingB, findingC},
			previousFingerprints: []finding.FingerprintID{fpA},
			wantNewCount:         2,
			wantExistingCount:    1,
			wantResolvedCount:    0,
		},
		{
			name:                 "resolved fingerprints counted",
			current:              []finding.Finding{findingA},
			previousFingerprints: []finding.FingerprintID{fpA, fpB, fpC},
			wantNewCount:         0,
			wantExistingCount:    1,
			wantResolvedCount:    2,
		},
		{
			name:                 "empty current means all previous resolved",
			current:              nil,
			previousFingerprints: []finding.FingerprintID{fpA, fpB},
			wantNewCount:         0,
			wantExistingCount:    0,
			wantResolvedCount:    2,
		},
		{
			name:                 "both empty yields empty result",
			current:              nil,
			previousFingerprints: nil,
			wantNewCount:         0,
			wantExistingCount:    0,
			wantResolvedCount:    0,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			result := tracking.ClassifyFindings(tcase.current, tcase.previousFingerprints)

			assert.Len(t, result.NewFindings(), tcase.wantNewCount)
			assert.Len(t, result.ExistingFindings(), tcase.wantExistingCount)
			assert.Equal(t, tcase.wantResolvedCount, result.ResolvedCount())
		})
	}
}

func TestTrackingResultDefensiveCopies(t *testing.T) {
	fp := mustFingerprintID(t, "fp-1")
	f := mustFinding(t, "tool", "rule", finding.SeverityError, "file.go", 1, "msg", fp)

	result := tracking.NewResult(
		[]finding.Finding{f},
		[]finding.Finding{f},
		0,
	)

	newFindings := result.NewFindings()
	newFindings[0] = mustFinding(t, "other", "other", finding.SeverityNote, "other.go", 99, "other", mustFingerprintID(t, "other"))

	assert.Equal(t, "tool", result.NewFindings()[0].Tool())

	existingFindings := result.ExistingFindings()
	existingFindings[0] = mustFinding(t, "other", "other", finding.SeverityNote, "other.go", 99, "other", mustFingerprintID(t, "other"))

	assert.Equal(t, "tool", result.ExistingFindings()[0].Tool())
}

func TestClassifyIdentifiers(t *testing.T) {
	tests := []struct {
		name              string
		current           []string
		previous          []string
		wantNewCount      int
		wantExistingCount int
		wantResolvedCount int
		wantNewIDs        map[string]bool
	}{
		{
			name:              "all new when no previous",
			current:           []string{"a", "b", "c"},
			previous:          nil,
			wantNewCount:      3,
			wantExistingCount: 0,
			wantResolvedCount: 0,
			wantNewIDs:        map[string]bool{"a": true, "b": true, "c": true},
		},
		{
			name:              "all new when previous empty",
			current:           []string{"a", "b"},
			previous:          []string{},
			wantNewCount:      2,
			wantExistingCount: 0,
			wantResolvedCount: 0,
			wantNewIDs:        map[string]bool{"a": true, "b": true},
		},
		{
			name:              "all resolved when current empty",
			current:           []string{},
			previous:          []string{"a", "b", "c"},
			wantNewCount:      0,
			wantExistingCount: 0,
			wantResolvedCount: 3,
			wantNewIDs:        map[string]bool{},
		},
		{
			name:              "all existing when identical",
			current:           []string{"a", "b", "c"},
			previous:          []string{"a", "b", "c"},
			wantNewCount:      0,
			wantExistingCount: 3,
			wantResolvedCount: 0,
			wantNewIDs:        map[string]bool{},
		},
		{
			name:              "mixed partition",
			current:           []string{"a", "c", "e", "f"},
			previous:          []string{"a", "b", "c", "d", "e"},
			wantNewCount:      1,
			wantExistingCount: 3,
			wantResolvedCount: 2,
			wantNewIDs:        map[string]bool{"f": true},
		},
		{
			name:              "both nil yields empty",
			current:           nil,
			previous:          nil,
			wantNewCount:      0,
			wantExistingCount: 0,
			wantResolvedCount: 0,
			wantNewIDs:        map[string]bool{},
		},
		{
			name:              "duplicates input current inflate counts when not deduplicated",
			current:           []string{"a", "b", "a"},
			previous:          []string{"a"},
			wantNewCount:      2,
			wantExistingCount: 1,
			wantResolvedCount: 0,
			wantNewIDs:        map[string]bool{"a": true, "b": true},
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			result := tracking.ClassifyIdentifiers(tcase.current, tcase.previous)

			assert.Equal(t, tcase.wantNewCount, result.NewCount())
			assert.Equal(t, tcase.wantExistingCount, result.ExistingCount())
			assert.Equal(t, tcase.wantResolvedCount, result.ResolvedCount())
			assert.Equal(t, tcase.wantNewIDs, result.NewIdentifiers())
		})
	}
}

func TestIdentifierClassificationDefensiveCopy(t *testing.T) {
	result := tracking.ClassifyIdentifiers([]string{"a", "b"}, nil)

	ids := result.NewIdentifiers()
	ids["injected"] = true

	assert.False(t, result.NewIdentifiers()["injected"])
}

func mustFingerprintID(t *testing.T, value string) finding.FingerprintID {
	t.Helper()
	fp, err := finding.NewFingerprintID(value)
	require.NoError(t, err)
	return fp
}

func mustFinding(t *testing.T, tool, ruleID string, severity finding.Severity, filePath string, line int, message string, fp finding.FingerprintID) finding.Finding {
	t.Helper()
	f, err := finding.NewFinding(tool, ruleID, severity, filePath, line, message, fp)
	require.NoError(t, err)
	return f
}
