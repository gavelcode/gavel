package finding_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
)

func mustFinding(t *testing.T, fingerprint string, line int) finding.Finding {
	t.Helper()
	fp, err := finding.NewFingerprintID(fingerprint)
	require.NoError(t, err)
	fnd, err := finding.NewFinding("golangci-lint", "varnamelen", finding.SeverityError, "a.go", line, "name too short", fp)
	require.NoError(t, err)
	return fnd
}

func mustContent(t *testing.T, findings ...finding.Finding) finding.Content {
	t.Helper()
	content, err := finding.NewContent(evidence.SubtypeCodeQuality, findings)
	require.NoError(t, err)
	return content
}

func TestContentMergeDeduplicatesByFingerprint(t *testing.T) {
	left := mustContent(t, mustFinding(t, "fp-dup", 10))
	right := mustContent(t, mustFinding(t, "fp-dup", 10))

	merged, err := left.Merge(right)

	require.NoError(t, err)
	fc, ok := merged.(finding.Content)
	require.True(t, ok)
	assert.Len(t, fc.Findings(), 1)
}

func TestContentMergeKeepsDistinctFindings(t *testing.T) {
	left := mustContent(t, mustFinding(t, "fp-a", 1))
	right := mustContent(t, mustFinding(t, "fp-b", 2))

	merged, err := left.Merge(right)

	require.NoError(t, err)
	fc, ok := merged.(finding.Content)
	require.True(t, ok)
	assert.Len(t, fc.Findings(), 2)
}
