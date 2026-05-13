package evidencedto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
)

func TestExtractFindings(t *testing.T) {
	evidences := []evidencedto.Evidence{
		{Findings: []evidencedto.Finding{{RuleID: "r1"}, {RuleID: "r2"}}},
		{Findings: []evidencedto.Finding{{RuleID: "r3"}}},
		{Subtype: "coverage"},
	}

	got := evidencedto.ExtractFindings(evidences)

	require.Len(t, got, 3)
	assert.Equal(t, "r1", got[0].RuleID)
	assert.Equal(t, "r3", got[2].RuleID)
}

func TestExtractFindings_Empty(t *testing.T) {
	got := evidencedto.ExtractFindings(nil)
	assert.Nil(t, got)
}

func TestExtractFindingsDeduplicatesByFingerprint(t *testing.T) {
	evidences := []evidencedto.Evidence{
		{Findings: []evidencedto.Finding{
			{RuleID: "errcheck", FingerprintID: "aaa", FilePath: "handler.go", Line: 10},
			{RuleID: "revive", FingerprintID: "bbb", FilePath: "handler.go", Line: 1},
		}},
		{Findings: []evidencedto.Finding{
			{RuleID: "errcheck", FingerprintID: "aaa", FilePath: "handler.go", Line: 10},
			{RuleID: "errcheck", FingerprintID: "ccc", FilePath: "handler_test.go", Line: 5},
		}},
	}

	got := evidencedto.ExtractFindings(evidences)

	require.Len(t, got, 3, "duplicate fingerprint 'aaa' must be collapsed")
	fingerprints := make([]string, len(got))
	for i, f := range got {
		fingerprints[i] = f.FingerprintID
	}
	assert.ElementsMatch(t, []string{"aaa", "bbb", "ccc"}, fingerprints)
}

func TestExtractFindingsPreservesOrderFirstSeen(t *testing.T) {
	evidences := []evidencedto.Evidence{
		{Findings: []evidencedto.Finding{
			{RuleID: "r1", FingerprintID: "fp1"},
			{RuleID: "r2", FingerprintID: "fp2"},
		}},
		{Findings: []evidencedto.Finding{
			{RuleID: "r1-dup", FingerprintID: "fp1"},
			{RuleID: "r3", FingerprintID: "fp3"},
		}},
	}

	got := evidencedto.ExtractFindings(evidences)

	require.Len(t, got, 3)
	assert.Equal(t, "r1", got[0].RuleID, "first-seen wins for duplicate fingerprint")
	assert.Equal(t, "r2", got[1].RuleID)
	assert.Equal(t, "r3", got[2].RuleID)
}

func TestExtractFingerprints(t *testing.T) {
	findings := []evidencedto.Finding{
		{FingerprintID: "fp1"},
		{FingerprintID: ""},
		{FingerprintID: "fp3"},
	}

	got := evidencedto.ExtractFingerprints(findings)

	assert.Equal(t, []string{"fp1", "fp3"}, got)
}

func TestExtractFingerprints_Empty(t *testing.T) {
	got := evidencedto.ExtractFingerprints(nil)
	assert.Empty(t, got)
}

func TestExtractViolations_NilEvidence(t *testing.T) {
	got := evidencedto.ExtractViolations(nil)
	assert.Nil(t, got)
}

func TestExtractViolations_NoArchitecture(t *testing.T) {
	ev := &evidencedto.Evidence{Subtype: "code_quality"}
	got := evidencedto.ExtractViolations(ev)
	assert.Nil(t, got)
}

func TestExtractViolations_WithViolations(t *testing.T) {
	evidence := &evidencedto.Evidence{
		Architecture: &evidencedto.Architecture{
			Violations: []evidencedto.Violation{
				{Rule: "LayerRule", SourcePkg: "api", TargetPkg: "auth"},
			},
		},
	}

	got := evidencedto.ExtractViolations(evidence)

	require.Len(t, got, 1)
	assert.Equal(t, "LayerRule", got[0].Rule)
}

func TestExtractArchIDs(t *testing.T) {
	violations := []evidencedto.Violation{
		{Rule: "layer_violation", SourcePkg: "com.api", TargetPkg: "com.domain"},
		{Rule: "dependency_rule", SourcePkg: "com.infra", TargetPkg: "com.app"},
	}

	got := evidencedto.ExtractArchIDs(violations)

	assert.Equal(t, []string{
		"layer_violation:com.api:com.domain",
		"dependency_rule:com.infra:com.app",
	}, got)
}

func TestExtractArchIDs_Empty(t *testing.T) {
	got := evidencedto.ExtractArchIDs(nil)
	assert.Empty(t, got)
}

func TestFilterNewViolations(t *testing.T) {
	evidence := &evidencedto.Evidence{
		Subtype: "architecture",
		Architecture: &evidencedto.Architecture{
			Violations: []evidencedto.Violation{
				{Rule: "layer", SourcePkg: "a", TargetPkg: "b", Message: "existing"},
				{Rule: "layer", SourcePkg: "c", TargetPkg: "d", Message: "new one"},
				{Rule: "dep", SourcePkg: "e", TargetPkg: "f", Message: "also new"},
			},
		},
	}
	newIDs := map[string]bool{
		"layer:c:d": true,
		"dep:e:f":   true,
	}

	got := evidencedto.FilterNewViolations(evidence, newIDs)

	require.NotNil(t, got.Architecture)
	assert.Len(t, got.Architecture.Violations, 2)
	assert.Equal(t, "new one", got.Architecture.Violations[0].Message)
	assert.Equal(t, "also new", got.Architecture.Violations[1].Message)
}

func TestFilterNewViolations_NilEvidence(t *testing.T) {
	got := evidencedto.FilterNewViolations(nil, map[string]bool{"x:y:z": true})
	assert.Nil(t, got)
}

func TestFilterNewViolations_NoArchitecture(t *testing.T) {
	ev := &evidencedto.Evidence{Subtype: "code_quality"}
	got := evidencedto.FilterNewViolations(ev, map[string]bool{"x:y:z": true})
	assert.Equal(t, ev, got)
}

func TestReplaceArchEvidence_Replaces(t *testing.T) {
	evidences := []evidencedto.Evidence{
		{Subtype: "code_quality"},
		{Subtype: "architecture", Architecture: &evidencedto.Architecture{Violations: []evidencedto.Violation{{Rule: "old"}}}},
	}
	newArch := &evidencedto.Evidence{
		Subtype:      "architecture",
		Architecture: &evidencedto.Architecture{Violations: []evidencedto.Violation{{Rule: "new"}}},
	}

	got := evidencedto.ReplaceArchEvidence(evidences, newArch)

	require.Len(t, got, 2)
	assert.Equal(t, "code_quality", got[0].Subtype)
	assert.Equal(t, "new", got[1].Architecture.Violations[0].Rule)
}

func TestReplaceArchEvidence_RemovesWhenNil(t *testing.T) {
	evidences := []evidencedto.Evidence{
		{Subtype: "code_quality"},
		{Subtype: "architecture"},
	}

	got := evidencedto.ReplaceArchEvidence(evidences, nil)

	require.Len(t, got, 1)
	assert.Equal(t, "code_quality", got[0].Subtype)
}

func TestReplaceArchEvidence_AppendsWhenNotPresent(t *testing.T) {
	evidences := []evidencedto.Evidence{
		{Subtype: "code_quality"},
	}
	newArch := &evidencedto.Evidence{
		Subtype:      "architecture",
		Architecture: &evidencedto.Architecture{Violations: []evidencedto.Violation{{Rule: "new"}}},
	}

	got := evidencedto.ReplaceArchEvidence(evidences, newArch)

	require.Len(t, got, 2)
	assert.Equal(t, "architecture", got[1].Subtype)
}
