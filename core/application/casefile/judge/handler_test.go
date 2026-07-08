package judge_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	"github.com/usegavel/gavel/core/application/casefile/judge"
	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/verdict"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

var testTenant = tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))

func TestHandlerExecuteRendersVerdict(t *testing.T) {
	caseFiles := newFakeCaseFileRepo()
	projects := newFakeProjectRepo()
	caseFile, project := seedPassingScenario(t, caseFiles, projects)

	handler := judge.NewHandler(caseFiles, projects)
	cmd := mustCommand(t, caseFile.ID().String(), nil)

	result, err := handler.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, verdict.OutcomePass.String(), result.Verdict.Outcome)
	assert.Equal(t, caseFile.ID().String(), result.CaseFileID)
	assert.True(t, project.ID().Equal(caseFile.ProjectID()))
	assert.Equal(t, 1, caseFiles.saveCalls, "case file persisted after judging")
}

func TestHandlerExecuteDrainsEventsAndPersists(t *testing.T) {
	caseFiles := newFakeCaseFileRepo()
	projects := newFakeProjectRepo()
	caseFile, _ := seedFailingScenario(t, caseFiles, projects)

	handler := judge.NewHandler(caseFiles, projects)
	cmd := mustCommand(t, caseFile.ID().String(), nil)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotEmpty(t, result.Events, "events returned to caller")
	persisted, err := caseFiles.FindByID(context.Background(), testTenant, caseFile.ID())
	require.NoError(t, err)
	assert.True(t, persisted.IsJudged(), "persisted case file carries verdict")
}

func TestHandlerExecuteFailingGate(t *testing.T) {
	caseFiles := newFakeCaseFileRepo()
	projects := newFakeProjectRepo()
	caseFile, _ := seedFailingScenario(t, caseFiles, projects)

	handler := judge.NewHandler(caseFiles, projects)
	cmd := mustCommand(t, caseFile.ID().String(), nil)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, verdict.OutcomeFail.String(), result.Verdict.Outcome)
}

func TestHandlerExecuteCaseFileNotFound(t *testing.T) {
	caseFiles := newFakeCaseFileRepo()
	projects := newFakeProjectRepo()

	handler := judge.NewHandler(caseFiles, projects)
	cmd := mustCommand(t, "missing", nil)

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
}

func TestHandlerExecuteProjectNotFound(t *testing.T) {
	caseFiles := newFakeCaseFileRepo()
	projects := newFakeProjectRepo()
	projectID := projectmodel.NewProjectID(uuid.New())
	caseFile, err := casefile.NewCaseFile(testTenant, projectID, "abc", "main", time.Now().UTC(), time.Now().UTC())
	require.NoError(t, err)
	caseFiles.seed(caseFile)

	handler := judge.NewHandler(caseFiles, projects)
	cmd := mustCommand(t, caseFile.ID().String(), nil)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
}

func TestHandlerExecutePropagatesAlreadyJudged(t *testing.T) {
	caseFiles := newFakeCaseFileRepo()
	projects := newFakeProjectRepo()
	caseFile, _ := seedPassingScenario(t, caseFiles, projects)

	handler := judge.NewHandler(caseFiles, projects)
	cmd := mustCommand(t, caseFile.ID().String(), nil)

	_, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, casefile.ErrAlreadyJudged)
}

func TestHandlerExecutePassesTracking(t *testing.T) {
	caseFiles := newFakeCaseFileRepo()
	projects := newFakeProjectRepo()
	caseFile, _ := seedFailingScenario(t, caseFiles, projects)

	tracking := evidencedto.Tracking{}

	handler := judge.NewHandler(caseFiles, projects)
	cmd := mustCommand(t, caseFile.ID().String(), &tracking)

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, verdict.OutcomePass.String(), result.Verdict.Outcome,
		"empty new-findings tracking filters all findings; gate passes")
}

func TestHandlerExecuteSaveErrorPropagated(t *testing.T) {
	caseFiles := newFakeCaseFileRepo()
	projects := newFakeProjectRepo()
	caseFile, _ := seedPassingScenario(t, caseFiles, projects)
	caseFiles.saveErr = errors.New("disk full")

	handler := judge.NewHandler(caseFiles, projects)
	cmd := mustCommand(t, caseFile.ID().String(), nil)

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
}

func TestNewHandlerRejectsNilRepos(t *testing.T) {
	assert.Panics(t, func() { judge.NewHandler(nil, newFakeProjectRepo()) })
	assert.Panics(t, func() { judge.NewHandler(newFakeCaseFileRepo(), nil) })
}

func TestHandlerExecuteFindByIDError(t *testing.T) {
	caseFiles := newFakeCaseFileRepo()
	projects := newFakeProjectRepo()
	caseFiles.findErr = errors.New("db connection lost")

	handler := judge.NewHandler(caseFiles, projects)
	cmd := mustCommand(t, "00000000-0000-0000-0000-000000000001", nil)

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load case file")
}

func TestHandlerExecuteTrackingConversionError(t *testing.T) {
	caseFiles := newFakeCaseFileRepo()
	projects := newFakeProjectRepo()
	caseFile, _ := seedPassingScenario(t, caseFiles, projects)

	tracking := &evidencedto.Tracking{
		NewFindings: []evidencedto.Finding{{Severity: "INVALID_SEVERITY"}},
	}

	handler := judge.NewHandler(caseFiles, projects)
	cmd, err := judge.NewCommand(testTenant.String(), caseFile.ID().String(), tracking)
	require.NoError(t, err)

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tracking")
}

func mustCommand(t *testing.T, id string, tracking *evidencedto.Tracking) judge.Command {
	t.Helper()
	cmd, err := judge.NewCommand(testTenant.String(), id, tracking)
	require.NoError(t, err)
	return cmd
}

func seedPassingScenario(t *testing.T, caseFiles *fakeCaseFileRepo, projects *fakeProjectRepo) (casefile.CaseFile, projectmodel.Project) {
	t.Helper()
	project := buildProject(t, qualitygate.Gate{})
	projects.seed(project)

	caseFile, err := casefile.NewCaseFile(testTenant, project.ID(), "abc", "main", time.Now().UTC(), time.Now().UTC())
	require.NoError(t, err)
	caseFiles.seed(caseFile)
	return caseFile, project
}

func seedFailingScenario(t *testing.T, caseFiles *fakeCaseFileRepo, projects *fakeProjectRepo) (casefile.CaseFile, projectmodel.Project) {
	t.Helper()
	rule, err := qualitygate.NewRule(
		evidence.SubtypeCodeQuality,
		qualitygate.NewZeroTolerance(),
	)
	require.NoError(t, err)
	qg, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	project := buildProject(t, qg)
	projects.seed(project)

	caseFile, err := casefile.NewCaseFile(testTenant, project.ID(), "abc", "main", time.Now().UTC(), time.Now().UTC())
	require.NoError(t, err)
	require.NoError(t, caseFile.AddEvidence(buildFindingsEvidence(t), time.Now().UTC()))
	caseFiles.seed(caseFile)
	return caseFile, project
}

func buildProject(t *testing.T, qg qualitygate.Gate) projectmodel.Project {
	t.Helper()
	p, err := projectmodel.NewProject(testTenant, "svc", "svc", "//svc/...")
	require.NoError(t, err)
	p.UpdateQualityGate(qg, time.Now().UTC())
	p.ClearEvents()
	return p
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
