package submit_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/classify"
	"github.com/usegavel/gavel/core/application/casefile/createcasefile"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/casefile/finalize"
	"github.com/usegavel/gavel/core/application/casefile/ingestevidence"
	"github.com/usegavel/gavel/core/application/casefile/judge"
	"github.com/usegavel/gavel/core/application/casefile/submit"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

func TestExecuteReturnsVerdictAndDelta(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t)
	project.UpdateBaseline("main", []string{"fp-existing", "fp-resolved"}, nil, nil, nil)
	projRepo.seed(project)

	handler := newSubmitHandler(cfRepo, projRepo)

	cmd := mustCommand(t, project.ID().String(), "abc123", "main",
		[]evidencedto.Evidence{findingsEvidence("fp-new", "fp-existing")},
		[]string{"fp-new", "fp-existing"},
		nil,
		finalize.ArchDeltaInput{},
		false,
	)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotEmpty(t, result.CaseFileID)
	assert.Equal(t, "pass", result.Verdict.Outcome)
	assert.True(t, result.Delta.HasPrevious)
	assert.Equal(t, 1, result.Delta.NewCount)
	assert.Equal(t, 1, result.Delta.FixedCount)
	assert.Equal(t, 1, result.Delta.ExistingCount)
}

func TestExecuteUpdatesBaselineOnPass(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t)
	projRepo.seed(project)

	handler := newSubmitHandler(cfRepo, projRepo)

	cmd := mustCommand(t, project.ID().String(), "abc123", "main",
		[]evidencedto.Evidence{findingsEvidence("fp-1")},
		[]string{"fp-1"},
		[]string{"arch-1"},
		finalize.ArchDeltaInput{},
		false,
	)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Verdict.Outcome)

	saved := projRepo.lastSaved()
	baseline := saved.Baseline("main")
	assert.True(t, baseline.HasPrevious())
	assert.Equal(t, []string{"fp-1"}, baseline.Fingerprints())
}

func TestExecuteSeedsBaselineWithFileCoverage(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t)
	projRepo.seed(project)

	handler := newSubmitHandler(cfRepo, projRepo)

	cmd, err := submit.NewCommand(project.ID().String(), "abc123", "main",
		[]evidencedto.Evidence{findingsEvidence("fp-1")},
		[]string{"fp-1"}, nil, finalize.ArchDeltaInput{},
		[]evidencedto.FileCoverage{{FilePath: "file.go", Covered: []int{1, 2}, Uncovered: []int{3}}},
		false, false, time.Now().UTC())
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	baseline := projRepo.lastSaved().Baseline("main")
	entries := baseline.FileCoverage()
	require.Len(t, entries, 1)
	assert.Equal(t, "file.go", entries[0].FilePath())
	assert.Equal(t, []int{1, 2}, entries[0].Covered())
	assert.Equal(t, []int{3}, entries[0].Uncovered())
}

func TestExecuteInvalidCommandRejected(t *testing.T) {
	_, err := submit.NewCommand("", "sha", "main", nil, nil, nil, finalize.ArchDeltaInput{}, nil, false, false, time.Now())
	assert.ErrorIs(t, err, submit.ErrInvalidCommand)
}

func TestExecuteCreateCaseFileCommandError(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t)
	projRepo.seed(project)

	handler := newSubmitHandler(cfRepo, projRepo)

	cmd, err := submit.NewCommand(project.ID().String(), "abc", "main",
		[]evidencedto.Evidence{findingsEvidence("fp-1")},
		[]string{"fp-1"}, nil, finalize.ArchDeltaInput{}, nil, false, false, time.Time{})
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create case file")
}

func TestExecuteCreateCaseFileError(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	handler := newSubmitHandler(cfRepo, projRepo)

	cmd := mustCommand(t, "00000000-0000-0000-0000-000000000001", "abc", "main",
		[]evidencedto.Evidence{findingsEvidence("fp-1")},
		[]string{"fp-1"},
		nil,
		finalize.ArchDeltaInput{},
		false,
	)

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create case file")
}

func TestExecuteIngestCommandError(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t)
	projRepo.seed(project)

	handler := newSubmitHandler(cfRepo, projRepo)

	cmd := mustCommand(t, project.ID().String(), "abc", "main",
		nil,
		nil,
		nil,
		finalize.ArchDeltaInput{},
		false,
	)

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ingest evidence")
}

func TestExecuteIngestExecuteError(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t)
	projRepo.seed(project)

	handler := newSubmitHandler(cfRepo, projRepo)

	cmd := mustCommand(t, project.ID().String(), "abc", "main",
		[]evidencedto.Evidence{{Subtype: "INVALID", Source: "test", CollectedAt: time.Now().UTC()}},
		nil,
		nil,
		finalize.ArchDeltaInput{},
		false,
	)

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ingest evidence")
}

func TestExecuteFinalizeError(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()

	project := mustProject(t)
	project.UpdateBaseline("main", []string{"fp-old"}, nil, nil, nil)
	projRepo.seed(project)

	cfRepo.fpErr = errors.New("db timeout")

	handler := newSubmitHandler(cfRepo, projRepo)

	cmd := mustCommand(t, project.ID().String(), "abc", "main",
		[]evidencedto.Evidence{findingsEvidence("fp-1")},
		[]string{"fp-1"},
		nil,
		finalize.ArchDeltaInput{},
		false,
	)

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "finalize")
}

func TestNewHandlerRejectsNilDeps(t *testing.T) {
	cfRepo := newFakeCaseFileRepo()
	projRepo := newFakeProjectRepo()
	classifyH := classify.NewHandler(cfRepo)
	judgeH := judge.NewHandler(cfRepo, projRepo)
	finalizeH := finalize.NewHandler(cfRepo, projRepo, classifyH, judgeH, nil)
	createH := createcasefile.NewHandler(cfRepo, projRepo)
	ingestH := ingestevidence.NewHandler(cfRepo)

	assert.Panics(t, func() { submit.NewHandler(nil, ingestH, finalizeH) })
	assert.Panics(t, func() { submit.NewHandler(createH, nil, finalizeH) })
	assert.Panics(t, func() { submit.NewHandler(createH, ingestH, nil) })
}

func newSubmitHandler(cfRepo *fakeCaseFileRepo, projRepo *fakeProjectRepo) *submit.Handler {
	classifyH := classify.NewHandler(cfRepo)
	judgeH := judge.NewHandler(cfRepo, projRepo)
	finalizeH := finalize.NewHandler(cfRepo, projRepo, classifyH, judgeH, nil)
	createH := createcasefile.NewHandler(cfRepo, projRepo)
	ingestH := ingestevidence.NewHandler(cfRepo)
	return submit.NewHandler(createH, ingestH, finalizeH)
}

func mustProject(t *testing.T) projectmodel.Project {
	t.Helper()
	p, err := projectmodel.NewProject("test", "test", "//test/...")
	require.NoError(t, err)
	return p
}

func mustCommand(t *testing.T, projectID, commitSHA, branch string, evidences []evidencedto.Evidence, fps, archIDs []string, archDelta finalize.ArchDeltaInput, quick bool) submit.Command {
	t.Helper()
	cmd, err := submit.NewCommand(projectID, commitSHA, branch, evidences, fps, archIDs, archDelta, nil, quick, false, time.Now().UTC())
	require.NoError(t, err)
	return cmd
}

func findingsEvidence(fingerprints ...string) evidencedto.Evidence {
	findings := make([]evidencedto.Finding, 0, len(fingerprints))
	for _, fingerprint := range fingerprints {
		findings = append(findings, evidencedto.Finding{
			Tool:          "test-tool",
			RuleID:        "rule1",
			Severity:      "warning",
			FilePath:      "file.go",
			Line:          1,
			Message:       "test finding",
			FingerprintID: fingerprint,
		})
	}
	return evidencedto.Evidence{
		Subtype:     "code_quality",
		Source:      "test",
		CollectedAt: time.Now().UTC(),
		Findings:    findings,
	}
}
