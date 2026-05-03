package evidence_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
)

func validFindingsContent(t *testing.T) finding.Content {
	t.Helper()
	fp, err := finding.NewFingerprintID("abc123")
	require.NoError(t, err)
	f, err := finding.NewFinding("pmd", "rule1", finding.SeverityWarning, "file.go", 1, "msg", fp)
	require.NoError(t, err)
	content, err := finding.NewContent(evidence.SubtypeCodeQuality, []finding.Finding{f})
	require.NoError(t, err)
	return content
}

func validCoverageContent(t *testing.T) coverage.Content {
	t.Helper()
	content, err := coverage.NewContent(100, 80, nil)
	require.NoError(t, err)
	return content
}

func TestNewEvidence(t *testing.T) {
	now := time.Now()
	findingsContent := validFindingsContent(t)
	coverageContent := validCoverageContent(t)

	tests := []struct {
		name        string
		subtype     evidence.Subtype
		source      string
		content     evidence.Content
		collectedAt time.Time
		wantErr     bool
	}{
		{
			name:        "shouldCreateValidEvidenceWithFindingsContent",
			subtype:     evidence.SubtypeCodeQuality,
			source:      "pmd",
			content:     findingsContent,
			collectedAt: now,
		},
		{
			name:        "shouldCreateValidEvidenceWithCoverageContent",
			subtype:     evidence.SubtypeCoverage,
			source:      "lcov",
			content:     coverageContent,
			collectedAt: now,
		},
		{
			name:        "shouldRejectEmptySource",
			subtype:     evidence.SubtypeCodeQuality,
			source:      "",
			content:     findingsContent,
			collectedAt: now,
			wantErr:     true,
		},
		{
			name:        "shouldRejectBlankSource",
			subtype:     evidence.SubtypeCodeQuality,
			source:      "   ",
			content:     findingsContent,
			collectedAt: now,
			wantErr:     true,
		},
		{
			name:        "shouldRejectNilContent",
			subtype:     evidence.SubtypeCodeQuality,
			source:      "pmd",
			content:     nil,
			collectedAt: now,
			wantErr:     true,
		},
		{
			name:        "shouldRejectZeroCollectedAt",
			subtype:     evidence.SubtypeCodeQuality,
			source:      "pmd",
			content:     findingsContent,
			collectedAt: time.Time{},
			wantErr:     true,
		},
		{
			name:        "shouldRejectMismatchedSubtypeAndContent",
			subtype:     evidence.SubtypeCoverage,
			source:      "pmd",
			content:     findingsContent,
			collectedAt: now,
			wantErr:     true,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			evid, err := evidence.NewEvidence(tcase.subtype, tcase.source, tcase.content, tcase.collectedAt)

			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, evidence.ErrInvalidEvidence)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tcase.subtype, evid.Subtype())
			assert.Equal(t, tcase.source, evid.Source())
			assert.Equal(t, tcase.content, evid.Content())
			assert.Equal(t, tcase.collectedAt, evid.CollectedAt())
			assert.Equal(t, tcase.subtype.Type(), evid.Type())
		})
	}
}

func TestNewEvidenceGeneratesUniqueIDs(t *testing.T) {
	now := time.Now()
	content := validFindingsContent(t)

	ev1, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, "pmd", content, now)
	require.NoError(t, err)
	ev2, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, "pmd", content, now)
	require.NoError(t, err)

	assert.False(t, ev1.ID().Equal(ev2.ID()))
}

func TestReconstituteEvidence(t *testing.T) {
	now := time.Now()
	findingsContent := validFindingsContent(t)

	t.Run("shouldPreserveProvidedID", func(t *testing.T) {
		existingID := evidence.NewEvidenceID(uuid.New())

		evid, err := evidence.ReconstituteEvidence(existingID, evidence.SubtypeCodeQuality, "pmd", findingsContent, now)

		require.NoError(t, err)
		assert.True(t, existingID.Equal(evid.ID()))
		assert.Equal(t, evidence.SubtypeCodeQuality, evid.Subtype())
		assert.Equal(t, "pmd", evid.Source())
		assert.Equal(t, findingsContent, evid.Content())
		assert.Equal(t, now, evid.CollectedAt())
		assert.Equal(t, evidence.SubtypeCodeQuality.Type(), evid.Type())
	})

	t.Run("shouldValidateSameInvariantsAsNew", func(t *testing.T) {
		existingID := evidence.NewEvidenceID(uuid.New())

		_, err := evidence.ReconstituteEvidence(existingID, evidence.SubtypeCodeQuality, "", findingsContent, now)

		require.Error(t, err)
		assert.ErrorIs(t, err, evidence.ErrInvalidEvidence)
	})
}
