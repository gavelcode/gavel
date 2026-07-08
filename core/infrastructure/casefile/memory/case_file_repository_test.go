package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"

	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/verdict"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/infrastructure/casefile/memory"
)

var testTenantID = tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))

func TestCaseFileRepositorySaveAndFindByID(t *testing.T) {
	repo := memory.NewCaseFileRepository()
	ctx := context.Background()

	cf := newTestCaseFile(t, "abc123", "main")
	require.NoError(t, repo.Save(ctx, cf))

	found, err := repo.FindByID(ctx, testTenantID, cf.ID())
	require.NoError(t, err)
	assert.Equal(t, cf.ID(), found.ID())
}

func TestCaseFileRepositoryFindByIDNotFound(t *testing.T) {
	repo := memory.NewCaseFileRepository()
	ctx := context.Background()

	id := casefile.NewCaseFileID(uuid.New())

	_, err := repo.FindByID(ctx, testTenantID, id)
	assert.ErrorIs(t, err, memory.ErrCaseFileNotFound)
}

func TestCaseFileRepositoryFindByProject(t *testing.T) {
	repo := memory.NewCaseFileRepository()
	ctx := context.Background()

	projectID := mustGenerateProjectID(t)
	cf1 := newCaseFileWithProject(t, projectID, "abc123", "main", time.Now().UTC())
	cf2 := newCaseFileWithProject(t, projectID, "def456", "main", time.Now().UTC())
	require.NoError(t, repo.Save(ctx, cf1))
	require.NoError(t, repo.Save(ctx, cf2))

	results, err := repo.FindByProject(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestCaseFileRepositoryFindLatestByBranch(t *testing.T) {
	repo := memory.NewCaseFileRepository()
	ctx := context.Background()

	projectID := mustGenerateProjectID(t)
	earlier := newCaseFileWithProject(t, projectID, "aaa", "main", time.Now().Add(-time.Hour))
	later := newCaseFileWithProject(t, projectID, "bbb", "main", time.Now())

	require.NoError(t, repo.Save(ctx, earlier))
	require.NoError(t, repo.Save(ctx, later))

	found, err := repo.FindLatestByBranch(ctx, projectID, "main")
	require.NoError(t, err)
	assert.Equal(t, later.ID(), found.ID())
}

func TestCaseFileRepositoryFindLatestByBranchNotFound(t *testing.T) {
	repo := memory.NewCaseFileRepository()
	ctx := context.Background()

	_, err := repo.FindLatestByBranch(ctx, mustGenerateProjectID(t), "main")
	assert.ErrorIs(t, err, memory.ErrCaseFileNotFound)
}

func TestCaseFileRepositoryFindFingerprintsByBranchEmpty(t *testing.T) {
	repo := memory.NewCaseFileRepository()
	ctx := context.Background()

	fps, err := repo.FindFingerprintIDsByBranch(ctx, mustGenerateProjectID(t), "main")
	require.NoError(t, err)
	assert.Nil(t, fps)
}

func TestCaseFileRepositorySaveUpdatesExisting(t *testing.T) {
	repo := memory.NewCaseFileRepository()
	ctx := context.Background()

	caseFile := newTestCaseFile(t, "abc123", "main")
	require.NoError(t, repo.Save(ctx, caseFile))

	ev := newTestEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	require.NoError(t, repo.Save(ctx, caseFile))

	found, err := repo.FindByID(ctx, testTenantID, caseFile.ID())
	require.NoError(t, err)
	assert.Len(t, found.Evidences(), 1)
}

func TestCaseFileRepositoryPreloadFingerprints(t *testing.T) {
	repo := memory.NewCaseFileRepository()
	ctx := context.Background()
	projectID := mustGenerateProjectID(t)

	fp1, err := finding.NewFingerprintID("fp-abc")
	require.NoError(t, err)
	fp2, err := finding.NewFingerprintID("fp-def")
	require.NoError(t, err)

	repo.PreloadFingerprints(projectID, "main", []finding.FingerprintID{fp1, fp2})

	fps, err := repo.FindFingerprintIDsByBranch(ctx, projectID, "main")
	require.NoError(t, err)
	assert.Len(t, fps, 2)
	assert.Equal(t, "fp-abc", fps[0].Value())
	assert.Equal(t, "fp-def", fps[1].Value())
}

func TestCaseFileRepositoryPreloadFingerprintsIgnoredWhenCaseFilesExist(t *testing.T) {
	repo := memory.NewCaseFileRepository()
	ctx := context.Background()
	projectID := mustGenerateProjectID(t)

	fp1, err := finding.NewFingerprintID("preloaded-fp")
	require.NoError(t, err)
	repo.PreloadFingerprints(projectID, "main", []finding.FingerprintID{fp1})

	caseFile := newCaseFileWithProject(t, projectID, "abc123", "main", time.Now().UTC())
	ev := newTestEvidence(t)
	require.NoError(t, caseFile.AddEvidence(ev, time.Now().UTC()))
	v, err := verdict.ReconstituteResult("pass", nil, time.Now().UTC())
	require.NoError(t, err)
	require.NoError(t, caseFile.RecordVerdict(v))
	require.NoError(t, repo.Save(ctx, caseFile))

	fps, err := repo.FindFingerprintIDsByBranch(ctx, projectID, "main")
	require.NoError(t, err)
	assert.Len(t, fps, 1)
	assert.Equal(t, "test-fp-1", fps[0].Value())
}

func TestCaseFileRepositoryFindFingerprintsByBranchDeduplicates(t *testing.T) {
	repo := memory.NewCaseFileRepository()
	ctx := context.Background()
	projectID := mustGenerateProjectID(t)

	fp, err := finding.NewFingerprintID("shared-fp")
	require.NoError(t, err)
	f1, err := finding.NewFinding("pmd", "rule1", finding.SeverityError, "a.java", 1, "msg1", fp)
	require.NoError(t, err)
	second, err := finding.NewFinding("pmd", "rule1", finding.SeverityError, "b.java", 2, "msg2", fp)
	require.NoError(t, err)

	fc1, err := finding.NewContent(evidence.SubtypeCodeQuality, []finding.Finding{f1})
	require.NoError(t, err)
	ev1, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, "pmd", fc1, time.Now().UTC())
	require.NoError(t, err)

	fc2, err := finding.NewContent(evidence.SubtypeCodeQuality, []finding.Finding{second})
	require.NoError(t, err)
	ev2, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, "pmd", fc2, time.Now().UTC())
	require.NoError(t, err)

	caseFile := newCaseFileWithProject(t, projectID, "dup-sha", "main", time.Now().UTC())
	require.NoError(t, caseFile.AddEvidence(ev1, time.Now().UTC()))
	require.NoError(t, caseFile.AddEvidence(ev2, time.Now().UTC()))
	v, err := verdict.ReconstituteResult("pass", nil, time.Now().UTC())
	require.NoError(t, err)
	require.NoError(t, caseFile.RecordVerdict(v))
	require.NoError(t, repo.Save(ctx, caseFile))

	fps, err := repo.FindFingerprintIDsByBranch(ctx, projectID, "main")
	require.NoError(t, err)
	assert.Len(t, fps, 1, "duplicate fingerprints must be deduplicated")
	assert.Equal(t, "shared-fp", fps[0].Value())
}

func TestCaseFileRepositoryFindFingerprintsByBranchSkipsNonFindingEvidence(t *testing.T) {
	repo := memory.NewCaseFileRepository()
	ctx := context.Background()
	projectID := mustGenerateProjectID(t)

	covLang, err := coverage.NewLanguage("go")
	require.NoError(t, err)
	langStats, err := coverage.NewLanguageStats(covLang, 10, 8)
	require.NoError(t, err)
	covContent, err := coverage.NewContent(10, 8, []coverage.LanguageStats{langStats})
	require.NoError(t, err)
	covEv, err := evidence.NewEvidence(evidence.SubtypeCoverage, "lcov", covContent, time.Now().UTC())
	require.NoError(t, err)

	findingFP, err := finding.NewFingerprintID("real-fp")
	require.NoError(t, err)
	f, err := finding.NewFinding("pmd", "rule1", finding.SeverityError, "a.java", 1, "msg", findingFP)
	require.NoError(t, err)
	fc, err := finding.NewContent(evidence.SubtypeCodeQuality, []finding.Finding{f})
	require.NoError(t, err)
	findingEv, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, "pmd", fc, time.Now().UTC())
	require.NoError(t, err)

	caseFile := newCaseFileWithProject(t, projectID, "mixed-sha", "main", time.Now().UTC())
	require.NoError(t, caseFile.AddEvidence(covEv, time.Now().UTC()))
	require.NoError(t, caseFile.AddEvidence(findingEv, time.Now().UTC()))
	v, err := verdict.ReconstituteResult("pass", nil, time.Now().UTC())
	require.NoError(t, err)
	require.NoError(t, caseFile.RecordVerdict(v))
	require.NoError(t, repo.Save(ctx, caseFile))

	fps, err := repo.FindFingerprintIDsByBranch(ctx, projectID, "main")
	require.NoError(t, err)
	assert.Len(t, fps, 1)
	assert.Equal(t, "real-fp", fps[0].Value())
}

func mustGenerateProjectID(t *testing.T) projectmodel.ProjectID {
	t.Helper()
	id := projectmodel.NewProjectID(uuid.New())
	return id
}

func newTestCaseFile(t *testing.T, commitSHA, branch string) casefile.CaseFile {
	t.Helper()
	return newCaseFileWithProject(t, mustGenerateProjectID(t), commitSHA, branch, time.Now().UTC())
}

func newCaseFileWithProject(t *testing.T, projectID projectmodel.ProjectID, commitSHA, branch string, startedAt time.Time) casefile.CaseFile {
	t.Helper()
	cf, err := casefile.NewCaseFile(testTenantID, projectID, commitSHA, branch, startedAt, startedAt)
	require.NoError(t, err)
	return cf
}

func newTestEvidence(t *testing.T) evidence.Evidence {
	t.Helper()
	fp, err := finding.NewFingerprintID("test-fp-1")
	require.NoError(t, err)
	f, err := finding.NewFinding("pmd", "rule1", finding.SeverityError, "file.java", 10, "msg", fp)
	require.NoError(t, err)
	fc, err := finding.NewContent(evidence.SubtypeCodeQuality, []finding.Finding{f})
	require.NoError(t, err)
	ev, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, "pmd", fc, time.Now().UTC())
	require.NoError(t, err)
	return ev
}
