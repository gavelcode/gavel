package evidencedto_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/architecture"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/license"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/toolexecution"
)

func TestFindingsEvidenceRoundTrip(t *testing.T) {
	original := buildFindingsEvidence(t)

	dto := evidencedto.EvidenceFromDomain(original)
	assert.Equal(t, original.ID().String(), dto.ID)
	assert.Equal(t, original.Subtype().String(), dto.Subtype)
	require.Len(t, dto.Findings, 1)
	assert.Equal(t, "spotbugs", dto.Findings[0].Tool)

	back, err := evidencedto.EvidenceToDomain(dto)
	require.NoError(t, err)
	assert.True(t, original.ID().Equal(back.ID()))
	fc, ok := back.Content().(finding.Content)
	require.True(t, ok)
	assert.Len(t, fc.Findings(), 1)
}

func TestCoverageEvidenceRoundTrip(t *testing.T) {
	original := buildCoverageEvidence(t)

	dto := evidencedto.EvidenceFromDomain(original)
	require.NotNil(t, dto.Coverage)
	assert.Equal(t, 100, dto.Coverage.TotalLines)
	assert.Equal(t, 80, dto.Coverage.CoveredLines)
	require.Len(t, dto.Coverage.ByLanguage, 1)
	assert.Equal(t, "go", dto.Coverage.ByLanguage[0].Language)

	back, err := evidencedto.EvidenceToDomain(dto)
	require.NoError(t, err)
	cc, ok := back.Content().(coverage.Content)
	require.True(t, ok)
	assert.Equal(t, 100, cc.TotalLines())
}

func TestLicenseEvidenceRoundTrip(t *testing.T) {
	original := buildLicenseEvidence(t)

	dto := evidencedto.EvidenceFromDomain(original)
	require.NotNil(t, dto.License)
	require.Len(t, dto.License.Dependencies, 1)
	assert.Equal(t, "lib", dto.License.Dependencies[0].Name)

	back, err := evidencedto.EvidenceToDomain(dto)
	require.NoError(t, err)
	lc, ok := back.Content().(license.Content)
	require.True(t, ok)
	assert.Len(t, lc.Dependencies(), 1)
}

func TestArchitectureEvidenceRoundTrip(t *testing.T) {
	original := buildArchitectureEvidence(t)

	dto := evidencedto.EvidenceFromDomain(original)
	require.NotNil(t, dto.Architecture)
	require.Len(t, dto.Architecture.Violations, 1)
	assert.Equal(t, "layer-dependency", dto.Architecture.Violations[0].Rule)
	assert.Equal(t, "domain/foo", dto.Architecture.Violations[0].SourcePkg)
	assert.Equal(t, "infra/bar", dto.Architecture.Violations[0].TargetPkg)

	back, err := evidencedto.EvidenceToDomain(dto)
	require.NoError(t, err)
	ac, ok := back.Content().(architecture.Content)
	require.True(t, ok)
	assert.Len(t, ac.Violations(), 1)
}

func TestToolExecutionEvidenceRoundTrip(t *testing.T) {
	original := buildToolExecutionEvidence(t)

	dto := evidencedto.EvidenceFromDomain(original)
	require.NotNil(t, dto.ToolExecution)
	require.Len(t, dto.ToolExecution.Failures, 2)
	assert.Equal(t, "pmd", dto.ToolExecution.Failures[0].Tool)
	assert.Equal(t, "exit code 1: analyzer crashed", dto.ToolExecution.Failures[0].Reason)

	back, err := evidencedto.EvidenceToDomain(dto)
	require.NoError(t, err)
	tec, ok := back.Content().(toolexecution.Content)
	require.True(t, ok)
	require.Len(t, tec.Failures(), 2)
	assert.Equal(t, "spotbugs", tec.Failures()[1].Tool())
	assert.Equal(t, "timed out after 300s", tec.Failures()[1].Reason())
}

func TestEvidenceToDomainToolExecutionMissingPayloadRejected(t *testing.T) {
	dto := evidencedto.Evidence{
		Subtype:     evidence.SubtypeToolExecution.String(),
		Source:      "sarif",
		CollectedAt: time.Now().UTC(),
	}
	_, err := evidencedto.EvidenceToDomain(dto)
	require.Error(t, err)
	assert.ErrorIs(t, err, evidencedto.ErrIncompatibleEvidence)
}

func TestEvidenceToDomainToolExecutionInvalidFailureRejected(t *testing.T) {
	dto := evidencedto.Evidence{
		Subtype:       evidence.SubtypeToolExecution.String(),
		Source:        "sarif",
		CollectedAt:   time.Now().UTC(),
		ToolExecution: &evidencedto.ToolExecution{Failures: []evidencedto.ToolFailure{{Tool: "", Reason: "x"}}},
	}
	_, err := evidencedto.EvidenceToDomain(dto)
	require.Error(t, err)
}

func TestEvidenceToDomainArchitectureMissingPayloadRejected(t *testing.T) {
	dto := evidencedto.Evidence{
		Subtype:     evidence.SubtypeArchitecture.String(),
		Source:      "archtest",
		CollectedAt: time.Now().UTC(),
	}
	_, err := evidencedto.EvidenceToDomain(dto)
	require.Error(t, err)
	assert.ErrorIs(t, err, evidencedto.ErrIncompatibleEvidence)
}

func TestEvidenceToDomainWithoutIDGeneratesFresh(t *testing.T) {
	dto := evidencedto.Evidence{
		Subtype:     evidence.SubtypeCoverage.String(),
		Source:      "go-test",
		CollectedAt: time.Now().UTC(),
		Coverage:    &evidencedto.Coverage{TotalLines: 10, CoveredLines: 5},
	}

	_, err := evidencedto.EvidenceToDomain(dto)
	require.NoError(t, err)
}

func TestEvidenceToDomainCoverageMissingPayloadRejected(t *testing.T) {
	dto := evidencedto.Evidence{
		Subtype:     evidence.SubtypeCoverage.String(),
		Source:      "go-test",
		CollectedAt: time.Now().UTC(),
	}
	_, err := evidencedto.EvidenceToDomain(dto)
	require.Error(t, err)
	assert.ErrorIs(t, err, evidencedto.ErrIncompatibleEvidence)
}

func TestEvidenceToDomainLicenseMissingPayloadRejected(t *testing.T) {
	dto := evidencedto.Evidence{
		Subtype:     evidence.SubtypeLicense.String(),
		Source:      "scanner",
		CollectedAt: time.Now().UTC(),
	}
	_, err := evidencedto.EvidenceToDomain(dto)
	require.Error(t, err)
	assert.ErrorIs(t, err, evidencedto.ErrIncompatibleEvidence)
}

func TestEvidenceToDomainInvalidSubtypeRejected(t *testing.T) {
	dto := evidencedto.Evidence{Subtype: "nonsense"}
	_, err := evidencedto.EvidenceToDomain(dto)
	require.Error(t, err)
}

func buildFindingsEvidence(t *testing.T) evidence.Evidence {
	t.Helper()
	fp, err := finding.NewFingerprintID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	require.NoError(t, err)
	f, err := finding.NewFinding("spotbugs", "rule1", finding.SeverityError, "file.go", 10, "msg", fp)
	require.NoError(t, err)
	content, err := finding.NewContent(evidence.SubtypeCodeQuality, []finding.Finding{f})
	require.NoError(t, err)
	ev, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, "spotbugs", content, time.Now().UTC())
	require.NoError(t, err)
	return ev
}

func buildCoverageEvidence(t *testing.T) evidence.Evidence {
	t.Helper()
	lang, err := coverage.NewLanguage("go")
	require.NoError(t, err)
	lc, err := coverage.NewLanguageStats(lang, 100, 80)
	require.NoError(t, err)
	content, err := coverage.NewContent(100, 80, []coverage.LanguageStats{lc})
	require.NoError(t, err)
	ev, err := evidence.NewEvidence(evidence.SubtypeCoverage, "go-test", content, time.Now().UTC())
	require.NoError(t, err)
	return ev
}

func buildArchitectureEvidence(t *testing.T) evidence.Evidence {
	t.Helper()
	av, err := architecture.NewViolation("layer-dependency", "domain/foo", "infra/bar", "domain imports infra")
	require.NoError(t, err)
	content, err := architecture.NewContent([]architecture.Violation{av})
	require.NoError(t, err)
	ev, err := evidence.NewEvidence(evidence.SubtypeArchitecture, "archtest", content, time.Now().UTC())
	require.NoError(t, err)
	return ev
}

func buildToolExecutionEvidence(t *testing.T) evidence.Evidence {
	t.Helper()
	f1, err := toolexecution.NewFailure("pmd", "exit code 1: analyzer crashed")
	require.NoError(t, err)
	f2, err := toolexecution.NewFailure("spotbugs", "timed out after 300s")
	require.NoError(t, err)
	content, err := toolexecution.NewContent([]toolexecution.Failure{f1, f2})
	require.NoError(t, err)
	ev, err := evidence.NewEvidence(evidence.SubtypeToolExecution, "sarif", content, time.Now().UTC())
	require.NoError(t, err)
	return ev
}

func buildLicenseEvidence(t *testing.T) evidence.Evidence {
	t.Helper()
	dl, err := license.NewDependency("lib", "1.0", "MIT")
	require.NoError(t, err)
	content, err := license.NewContent([]license.Dependency{dl})
	require.NoError(t, err)
	ev, err := evidence.NewEvidence(evidence.SubtypeLicense, "scanner", content, time.Now().UTC())
	require.NoError(t, err)
	return ev
}
