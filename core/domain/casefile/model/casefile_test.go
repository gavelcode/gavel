package model_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/toolexecution"
	"github.com/usegavel/gavel/core/domain/casefile/model/tracking"
	"github.com/usegavel/gavel/core/domain/casefile/model/verdict"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

var (
	testTime   = time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	testTenant = tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))
)

func mustProjectID(t *testing.T) projectmodel.ProjectID {
	t.Helper()
	id := projectmodel.NewProjectID(uuid.New())
	return id
}

func TestNewCaseFile(t *testing.T) {
	validProjectID := mustProjectID(t)

	tests := []struct {
		name      string
		projectID projectmodel.ProjectID
		commitSHA string
		branch    string
		startedAt time.Time
		createdAt time.Time
		wantErr   bool
	}{
		{
			name:      "valid case file",
			projectID: validProjectID,
			commitSHA: "abc123",
			branch:    "main",
			startedAt: testTime,
			createdAt: testTime,
		},
		{
			name:      "empty commitSHA rejected",
			projectID: validProjectID,
			commitSHA: "",
			branch:    "main",
			startedAt: testTime,
			createdAt: testTime,
			wantErr:   true,
		},
		{
			name:      "whitespace commitSHA rejected",
			projectID: validProjectID,
			commitSHA: "   ",
			branch:    "main",
			startedAt: testTime,
			createdAt: testTime,
			wantErr:   true,
		},
		{
			name:      "empty branch rejected",
			projectID: validProjectID,
			commitSHA: "abc123",
			branch:    "",
			startedAt: testTime,
			createdAt: testTime,
			wantErr:   true,
		},
		{
			name:      "zero startedAt rejected",
			projectID: validProjectID,
			commitSHA: "abc123",
			branch:    "main",
			startedAt: time.Time{},
			createdAt: testTime,
			wantErr:   true,
		},
		{
			name:      "zero createdAt rejected",
			projectID: validProjectID,
			commitSHA: "abc123",
			branch:    "main",
			startedAt: testTime,
			createdAt: time.Time{},
			wantErr:   true,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			caseF, err := model.NewCaseFile(testTenant, tcase.projectID, tcase.commitSHA, tcase.branch, tcase.startedAt, tcase.createdAt)

			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, model.ErrInvalidCaseFile)
				return
			}
			require.NoError(t, err)
			assert.True(t, testTenant.Equal(caseF.TenantID()))
			assert.True(t, tcase.projectID.Equal(caseF.ProjectID()))
			assert.Equal(t, tcase.commitSHA, caseF.CommitSHA())
			assert.Equal(t, tcase.branch, caseF.Branch())
			assert.Equal(t, tcase.startedAt, caseF.StartedAt())
			assert.Empty(t, caseF.Evidences())
			_, ok := caseF.Verdict()
			assert.False(t, ok)
			assert.False(t, caseF.IsJudged())
		})
	}
}

func TestCaseFileOpenedEventUsesCreatedAtNotStartedAt(t *testing.T) {
	startedAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)

	caseF, err := model.NewCaseFile(testTenant, mustProjectID(t), "abc123", "main", startedAt, createdAt)
	require.NoError(t, err)

	events := caseF.Events()
	require.Len(t, events, 1)

	opened, ok := events[0].(model.CaseFileOpened)
	require.True(t, ok)
	assert.Equal(t, createdAt, opened.OccurredAt())
	assert.NotEqual(t, startedAt, opened.OccurredAt())
}

func TestReconstituteCaseFile(t *testing.T) {
	validID := model.NewCaseFileID(uuid.New())
	validProjectID := mustProjectID(t)
	composed, err := verdict.Compose(nil, testTime)
	require.NoError(t, err)
	existingVerdict := verdictPtr(composed)

	tests := []struct {
		name      string
		id        model.CaseFileID
		projectID projectmodel.ProjectID
		commitSHA string
		branch    string
		startedAt time.Time
		evidences []evidence.Evidence
		verdict   *verdict.Result
		wantErr   bool
	}{
		{
			name:      "valid reconstitution with verdict",
			id:        validID,
			projectID: validProjectID,
			commitSHA: "abc123",
			branch:    "main",
			startedAt: testTime,
			verdict:   existingVerdict,
		},
		{
			name:      "valid reconstitution without verdict",
			id:        validID,
			projectID: validProjectID,
			commitSHA: "abc123",
			branch:    "main",
			startedAt: testTime,
		},
		{
			name:      "empty commitSHA rejected",
			id:        validID,
			projectID: validProjectID,
			commitSHA: "",
			branch:    "main",
			startedAt: testTime,
			wantErr:   true,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			caseF, err := model.ReconstituteCaseFile(tcase.id, testTenant, tcase.projectID, tcase.commitSHA, tcase.branch, tcase.startedAt, tcase.evidences, tcase.verdict, false)

			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, model.ErrInvalidCaseFile)
				return
			}
			require.NoError(t, err)
			assert.True(t, tcase.id.Equal(caseF.ID()))
			assert.True(t, testTenant.Equal(caseF.TenantID()))
			assert.True(t, tcase.projectID.Equal(caseF.ProjectID()))
			assert.Equal(t, tcase.commitSHA, caseF.CommitSHA())
			assert.Equal(t, tcase.branch, caseF.Branch())
			assert.Equal(t, tcase.startedAt, caseF.StartedAt())

			_, hasVerdict := caseF.Verdict()
			if tcase.verdict != nil {
				assert.True(t, hasVerdict)
				assert.True(t, caseF.IsJudged())
			} else {
				assert.False(t, hasVerdict)
				assert.False(t, caseF.IsJudged())
			}
		})
	}
}

func TestCaseFileIsFreshEvaluationDefaultsFalse(t *testing.T) {
	caseF := mustNewCaseFile(t)
	assert.False(t, caseF.IsFreshEvaluation())
}

func TestCaseFileMarkFreshEvaluation(t *testing.T) {
	caseF := mustNewCaseFile(t)
	caseF.MarkFreshEvaluation()
	assert.True(t, caseF.IsFreshEvaluation())
}

func TestReconstituteCaseFilePreservesFreshEvaluation(t *testing.T) {
	caseFileID := model.NewCaseFileID(uuid.New())

	caseF, err := model.ReconstituteCaseFile(caseFileID, testTenant, mustProjectID(t), "abc123", "main", testTime, nil, nil, true)
	require.NoError(t, err)
	assert.True(t, caseF.IsFreshEvaluation())

	cf2, err := model.ReconstituteCaseFile(caseFileID, testTenant, mustProjectID(t), "abc123", "main", testTime, nil, nil, false)
	require.NoError(t, err)
	assert.False(t, cf2.IsFreshEvaluation())
}

func TestCaseFileAddEvidence(t *testing.T) {
	caseF := mustNewCaseFile(t)
	evid := mustCodeQualityEvidence(t, "pmd", testTime)

	err := caseF.AddEvidence(evid, testTime)

	require.NoError(t, err)
	assert.Len(t, caseF.Evidences(), 1)
	assert.True(t, evid.ID().Equal(caseF.Evidences()[0].ID()))

	events := caseF.Events()
	require.Len(t, events, 2, "CaseFileOpened (creation) + EvidenceCollected")
	_, openedOK := events[0].(model.CaseFileOpened)
	require.True(t, openedOK, "first event is creation marker")
	collected, ok := events[1].(model.EvidenceCollected)
	require.True(t, ok)
	assert.True(t, caseF.ID().Equal(collected.CaseFileID()))
	assert.True(t, caseF.ProjectID().Equal(collected.ProjectID()))
	assert.Equal(t, evidence.SubtypeCodeQuality.String(), collected.Subtype())
	assert.Equal(t, "pmd", collected.Source())
}

func TestCaseFileAddEvidenceAfterJudge(t *testing.T) {
	caseF := mustNewCaseFile(t)
	evid := mustCodeQualityEvidence(t, "pmd", testTime)
	require.NoError(t, caseF.AddEvidence(evid, testTime))

	qGate := mustQualityGate(t, qualitygate.NewZeroTolerance())
	_, err := caseF.Judge(qGate, nil, testTime, nil)
	require.NoError(t, err)

	err = caseF.AddEvidence(evid, testTime)

	assert.ErrorIs(t, err, model.ErrAlreadyJudged)
}

func TestCaseFileJudgeSimplePass(t *testing.T) {
	caseF := mustNewCaseFile(t)
	evid := mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, nil)
	require.NoError(t, caseF.AddEvidence(evid, testTime))

	qGate := mustQualityGate(t, qualitygate.NewZeroTolerance())

	verdict, err := caseF.Judge(qGate, nil, testTime, nil)

	require.NoError(t, err)
	assert.Equal(t, "pass", verdict.Outcome().String())
	assert.True(t, caseF.IsJudged())
	_, hasVerdict := caseF.Verdict()
	assert.True(t, hasVerdict)
}

func TestCaseFileJudgeSimpleFail(t *testing.T) {
	caseF := mustNewCaseFile(t)
	findings := []finding.Finding{
		mustFinding(t, "pmd", "rule1", finding.SeverityError, "file.go", 1, "bad code", mustFingerprintID(t, "fp-1")),
	}
	evid := mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, findings)
	require.NoError(t, caseF.AddEvidence(evid, testTime))

	qGate := mustQualityGate(t, qualitygate.NewZeroTolerance())

	verdict, err := caseF.Judge(qGate, nil, testTime, nil)

	require.NoError(t, err)
	assert.Equal(t, "fail", verdict.Outcome().String())
}

func TestCaseFileJudgeFailsOnToolExecutionFailure(t *testing.T) {
	caseF := mustNewCaseFile(t)
	require.NoError(t, caseF.AddEvidence(mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, nil), testTime))
	failed := mustFailure(t, "golangci-lint", "compilation errors prevented analysis")
	require.NoError(t, caseF.AddEvidence(mustToolExecutionEvidence(t, "golangci-lint", testTime, []toolexecution.Failure{failed}), testTime))

	qGate := mustQualityGate(t, qualitygate.NewZeroTolerance())

	result, err := caseF.Judge(qGate, nil, testTime, nil)

	require.NoError(t, err)
	assert.Equal(t, "fail", result.Outcome().String())
	ruling := findRuling(t, result, evidence.SubtypeToolExecution)
	assert.False(t, ruling.Passed())
	assert.Contains(t, ruling.Detail(), "golangci-lint")
}

func TestCaseFileJudgeToolExecutionRulingAlwaysPresent(t *testing.T) {
	caseF := mustNewCaseFile(t)
	require.NoError(t, caseF.AddEvidence(mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, nil), testTime))

	qGate := mustQualityGate(t, qualitygate.NewZeroTolerance())

	result, err := caseF.Judge(qGate, nil, testTime, nil)

	require.NoError(t, err)
	assert.Equal(t, "pass", result.Outcome().String())
	ruling := findRuling(t, result, evidence.SubtypeToolExecution)
	assert.True(t, ruling.Passed())
}

func TestCaseFileJudgeDoesNotFailOnDegradedToolExecution(t *testing.T) {
	caseF := mustNewCaseFile(t)
	require.NoError(t, caseF.AddEvidence(mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, nil), testTime))
	degraded := mustDegradedFailure(t, "ErrorProne", "13 javac compilation error(s); results are incomplete")
	require.NoError(t, caseF.AddEvidence(mustToolExecutionEvidence(t, "ErrorProne", testTime, []toolexecution.Failure{degraded}), testTime))

	qGate := mustQualityGate(t, qualitygate.NewZeroTolerance())

	result, err := caseF.Judge(qGate, nil, testTime, nil)

	require.NoError(t, err)
	assert.Equal(t, "pass", result.Outcome().String(), "an incomplete analysis must not fail the verdict")
	ruling := findRuling(t, result, evidence.SubtypeToolExecution)
	assert.True(t, ruling.Passed())
	assert.Contains(t, ruling.Detail(), "incomplete analysis")
	assert.Contains(t, ruling.Detail(), "ErrorProne")
}

func TestCaseFileJudgeFailsWhenHardFailureAccompaniesDegraded(t *testing.T) {
	caseF := mustNewCaseFile(t)
	require.NoError(t, caseF.AddEvidence(mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, nil), testTime))
	failures := []toolexecution.Failure{
		mustDegradedFailure(t, "ErrorProne", "results are incomplete"),
		mustFailure(t, "pmd", "could not launch"),
	}
	require.NoError(t, caseF.AddEvidence(mustToolExecutionEvidence(t, "java", testTime, failures), testTime))

	qGate := mustQualityGate(t, qualitygate.NewZeroTolerance())

	result, err := caseF.Judge(qGate, nil, testTime, nil)

	require.NoError(t, err)
	assert.Equal(t, "fail", result.Outcome().String(), "a genuine hard failure must still fail even when a degraded run is present")
	ruling := findRuling(t, result, evidence.SubtypeToolExecution)
	assert.False(t, ruling.Passed())
	assert.Contains(t, ruling.Detail(), "pmd")
}

func mustFailure(t *testing.T, tool, reason string) toolexecution.Failure {
	t.Helper()
	failed, err := toolexecution.NewFailure(tool, reason)
	require.NoError(t, err)
	return failed
}

func mustDegradedFailure(t *testing.T, tool, reason string) toolexecution.Failure {
	t.Helper()
	failed, err := toolexecution.NewDegradedFailure(tool, reason)
	require.NoError(t, err)
	return failed
}

func mustToolExecutionEvidence(t *testing.T, source string, collectedAt time.Time, failures []toolexecution.Failure) evidence.Evidence {
	t.Helper()
	content, err := toolexecution.NewContent(failures)
	require.NoError(t, err)
	evid, err := evidence.NewEvidence(evidence.SubtypeToolExecution, source, content, collectedAt)
	require.NoError(t, err)
	return evid
}

func findRuling(t *testing.T, result verdict.Result, subtype evidence.Subtype) verdict.Ruling {
	t.Helper()
	for _, ruling := range result.Rulings() {
		if ruling.Subtype().Equal(subtype) {
			return ruling
		}
	}
	t.Fatalf("no ruling for subtype %s", subtype)
	return verdict.Ruling{}
}

func TestCaseFileJudgeTwice(t *testing.T) {
	caseF := mustNewCaseFile(t)
	evid := mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, nil)
	require.NoError(t, caseF.AddEvidence(evid, testTime))

	qGate := mustQualityGate(t, qualitygate.NewZeroTolerance())
	_, err := caseF.Judge(qGate, nil, testTime, nil)
	require.NoError(t, err)

	_, err = caseF.Judge(qGate, nil, testTime, nil)

	assert.ErrorIs(t, err, model.ErrAlreadyJudged)
}

func TestCaseFileJudgeMergesMultipleEvidencesSameSubtype(t *testing.T) {
	caseF := mustNewCaseFile(t)

	findings1 := []finding.Finding{
		mustFinding(t, "pmd", "rule1", finding.SeverityError, "a.go", 1, "err1", mustFingerprintID(t, "fp-1")),
	}
	findings2 := []finding.Finding{
		mustFinding(t, "spotbugs", "rule2", finding.SeverityError, "b.go", 2, "err2", mustFingerprintID(t, "fp-2")),
	}
	ev1 := mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, findings1)
	ev2 := mustCodeQualityEvidenceWithFindings(t, "spotbugs", testTime, findings2)
	require.NoError(t, caseF.AddEvidence(ev1, testTime))
	require.NoError(t, caseF.AddEvidence(ev2, testTime))

	strategy, err := qualitygate.NewCountBySeverity(1, 100, 100)
	require.NoError(t, err)
	qGate := mustQualityGate(t, strategy)

	verdict, err := caseF.Judge(qGate, nil, testTime, nil)

	require.NoError(t, err)
	assert.Equal(t, "fail", verdict.Outcome().String())
}

func TestCaseFileJudgeWithTracking(t *testing.T) {
	caseF := mustNewCaseFile(t)

	fpNew := mustFingerprintID(t, "fp-new")
	fpExisting := mustFingerprintID(t, "fp-existing")

	findings := []finding.Finding{
		mustFinding(t, "pmd", "rule1", finding.SeverityError, "a.go", 1, "new issue", fpNew),
		mustFinding(t, "pmd", "rule2", finding.SeverityError, "b.go", 2, "old issue", fpExisting),
	}
	evid := mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, findings)
	require.NoError(t, caseF.AddEvidence(evid, testTime))

	tracking := tracking.NewResult(
		[]finding.Finding{findings[0]},
		[]finding.Finding{findings[1]},
		0,
	)

	qGate := mustQualityGate(t, qualitygate.NewZeroTolerance())

	verdict, err := caseF.Judge(qGate, &tracking, testTime, nil)

	require.NoError(t, err)
	assert.Equal(t, "fail", verdict.Outcome().String())
}

func TestCaseFileJudgeWithoutTracking(t *testing.T) {
	caseF := mustNewCaseFile(t)

	findings := []finding.Finding{
		mustFinding(t, "pmd", "rule1", finding.SeverityError, "a.go", 1, "issue", mustFingerprintID(t, "fp-1")),
	}
	evid := mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, findings)
	require.NoError(t, caseF.AddEvidence(evid, testTime))

	qGate := mustQualityGate(t, qualitygate.NewZeroTolerance())

	verdict, err := caseF.Judge(qGate, nil, testTime, nil)

	require.NoError(t, err)
	assert.Equal(t, "fail", verdict.Outcome().String())
}

func TestCaseFileJudgeRejectsZeroEvaluatedAt(t *testing.T) {
	caseF := mustNewCaseFile(t)
	evid := mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, nil)
	require.NoError(t, caseF.AddEvidence(evid, testTime))

	qGate := mustQualityGate(t, qualitygate.NewZeroTolerance())

	_, err := caseF.Judge(qGate, nil, time.Time{}, nil)
	require.Error(t, err)
}

func TestCaseFileOpenedEventGetters(t *testing.T) {
	caseF, err := model.NewCaseFile(testTenant, mustProjectID(t), "abc123", "main", testTime, testTime)
	require.NoError(t, err)

	events := caseF.Events()
	require.Len(t, events, 1)
	opened := events[0].(model.CaseFileOpened)

	assert.True(t, caseF.ID().Equal(opened.CaseFileID()))
	assert.True(t, caseF.ProjectID().Equal(opened.ProjectID()))
	assert.Equal(t, "abc123", opened.CommitSHA())
	assert.Equal(t, "main", opened.Branch())
}

func TestCaseFileJudgeEmitsEvents(t *testing.T) {
	t.Run("VerdictRendered on pass", func(t *testing.T) {
		caseF := mustNewCaseFile(t)
		evid := mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, nil)
		require.NoError(t, caseF.AddEvidence(evid, testTime))
		caseF.ClearEvents()

		qGate := mustQualityGate(t, qualitygate.NewZeroTolerance())
		_, err := caseF.Judge(qGate, nil, testTime, nil)
		require.NoError(t, err)

		events := caseF.Events()
		require.Len(t, events, 1)
		rendered, ok := events[0].(model.VerdictRendered)
		require.True(t, ok)
		assert.True(t, caseF.ID().Equal(rendered.CaseFileID()))
		assert.True(t, caseF.ProjectID().Equal(rendered.ProjectID()))
		assert.Equal(t, "pass", rendered.Outcome())
	})

	t.Run("VerdictRendered and QualityGateFailed on fail", func(t *testing.T) {
		caseF := mustNewCaseFile(t)
		findings := []finding.Finding{
			mustFinding(t, "pmd", "rule1", finding.SeverityError, "a.go", 1, "bad", mustFingerprintID(t, "fp-1")),
		}
		evid := mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, findings)
		require.NoError(t, caseF.AddEvidence(evid, testTime))
		caseF.ClearEvents()

		qGate := mustQualityGate(t, qualitygate.NewZeroTolerance())
		_, err := caseF.Judge(qGate, nil, testTime, nil)
		require.NoError(t, err)

		events := caseF.Events()
		require.Len(t, events, 2)

		rendered, ok := events[0].(model.VerdictRendered)
		require.True(t, ok)
		assert.Equal(t, "fail", rendered.Outcome())

		failed, ok := events[1].(model.QualityGateFailed)
		require.True(t, ok)
		assert.True(t, caseF.ID().Equal(failed.CaseFileID()))
		assert.True(t, caseF.ProjectID().Equal(failed.ProjectID()))
		assert.Contains(t, failed.FailingSubtypes(), evidence.SubtypeCodeQuality.String())
	})
}

func TestCaseFileJudgeNoEvidenceForRule(t *testing.T) {
	caseF := mustNewCaseFile(t)

	rule, err := qualitygate.NewRule(
		evidence.SubtypeSAST,
		qualitygate.NewZeroTolerance(),
	)
	require.NoError(t, err)
	qGate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	verdict, err := caseF.Judge(qGate, nil, testTime, nil)

	require.NoError(t, err)
	assert.Equal(t, "pass", verdict.Outcome().String())
}

func TestCaseFileJudgeMinResolvedPasses(t *testing.T) {
	caseF := mustNewCaseFile(t)
	evid := mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, nil)
	require.NoError(t, caseF.AddEvidence(evid, testTime))

	rule, err := qualitygate.NewRule(
		evidence.SubtypeCodeQuality,
		qualitygate.NewZeroTolerance(),
		qualitygate.WithMinResolved(2),
	)
	require.NoError(t, err)
	qGate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	delta := &model.DeltaInput{
		FindingsResolved: 3,
	}

	v, err := caseF.Judge(qGate, nil, testTime, delta)
	require.NoError(t, err)
	assert.Equal(t, "pass", v.Outcome().String())
}

func TestCaseFileJudgeMinResolvedFails(t *testing.T) {
	caseF := mustNewCaseFile(t)
	evid := mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, nil)
	require.NoError(t, caseF.AddEvidence(evid, testTime))

	rule, err := qualitygate.NewRule(
		evidence.SubtypeCodeQuality,
		qualitygate.NewZeroTolerance(),
		qualitygate.WithMinResolved(5),
	)
	require.NoError(t, err)
	qGate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	delta := &model.DeltaInput{
		FindingsResolved: 2,
	}

	v, err := caseF.Judge(qGate, nil, testTime, delta)
	require.NoError(t, err)
	assert.Equal(t, "fail", v.Outcome().String())

	rulings := v.Rulings()
	require.Len(t, rulings, 2)
	assert.Contains(t, rulings[0].Detail(), "resolved 2")
	assert.Contains(t, rulings[0].Detail(), "min 5")
}

func TestCaseFileJudgeArchMinResolvedFails(t *testing.T) {
	caseF := mustNewCaseFile(t)

	rule, err := qualitygate.NewRule(
		evidence.SubtypeArchitecture,
		mustMaxViolations(t, 0),
		qualitygate.WithMinResolved(3),
	)
	require.NoError(t, err)
	qGate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	delta := &model.DeltaInput{
		ArchResolved: 1,
	}

	v, err := caseF.Judge(qGate, nil, testTime, delta)
	require.NoError(t, err)
	assert.Equal(t, "fail", v.Outcome().String())
	assert.Contains(t, v.Rulings()[0].Detail(), "resolved 1")
}

func TestCaseFileJudgeCoverageMinDeltaPasses(t *testing.T) {
	caseF := mustNewCaseFile(t)
	covEv := mustCoverageEvidence(t, 100, 85, testTime)
	require.NoError(t, caseF.AddEvidence(covEv, testTime))

	pct, err := qualitygate.NewMinPercentage(80)
	require.NoError(t, err)
	rule, err := qualitygate.NewRule(
		evidence.SubtypeCoverage,
		pct,
		qualitygate.WithMinDelta(0),
	)
	require.NoError(t, err)
	qGate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	prev := 80.0
	delta := &model.DeltaInput{
		PreviousCoverage: &prev,
		CurrentCoverage:  85.0,
	}

	v, err := caseF.Judge(qGate, nil, testTime, delta)
	require.NoError(t, err)
	assert.Equal(t, "pass", v.Outcome().String())
}

func TestCaseFileJudgeCoverageMinDeltaFails(t *testing.T) {
	caseF := mustNewCaseFile(t)
	covEv := mustCoverageEvidence(t, 100, 75, testTime)
	require.NoError(t, caseF.AddEvidence(covEv, testTime))

	pct, err := qualitygate.NewMinPercentage(70)
	require.NoError(t, err)
	rule, err := qualitygate.NewRule(
		evidence.SubtypeCoverage,
		pct,
		qualitygate.WithMinDelta(0),
	)
	require.NoError(t, err)
	qGate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	prev := 80.0
	delta := &model.DeltaInput{
		PreviousCoverage: &prev,
		CurrentCoverage:  75.0,
	}

	v, err := caseF.Judge(qGate, nil, testTime, delta)
	require.NoError(t, err)
	assert.Equal(t, "fail", v.Outcome().String())
	assert.Contains(t, v.Rulings()[0].Detail(), "coverage delta")
}

func TestCaseFileJudgeCoverageNoPreviousPassesDelta(t *testing.T) {
	caseF := mustNewCaseFile(t)
	covEv := mustCoverageEvidence(t, 100, 85, testTime)
	require.NoError(t, caseF.AddEvidence(covEv, testTime))

	pct, err := qualitygate.NewMinPercentage(80)
	require.NoError(t, err)
	rule, err := qualitygate.NewRule(
		evidence.SubtypeCoverage,
		pct,
		qualitygate.WithMinDelta(0),
	)
	require.NoError(t, err)
	qGate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	delta := &model.DeltaInput{
		CurrentCoverage: 85.0,
	}

	v, err := caseF.Judge(qGate, nil, testTime, delta)
	require.NoError(t, err)
	assert.Equal(t, "pass", v.Outcome().String())
}

func TestCaseFileJudgeCompoundFailAbsoluteAndDelta(t *testing.T) {
	caseF := mustNewCaseFile(t)
	findings := []finding.Finding{
		mustFinding(t, "pmd", "rule1", finding.SeverityError, "file.go", 1, "bad", mustFingerprintID(t, "fp-1")),
	}
	evid := mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, findings)
	require.NoError(t, caseF.AddEvidence(evid, testTime))

	rule, err := qualitygate.NewRule(
		evidence.SubtypeCodeQuality,
		qualitygate.NewZeroTolerance(),
		qualitygate.WithMinResolved(5),
	)
	require.NoError(t, err)
	qGate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	delta := &model.DeltaInput{
		FindingsResolved: 1,
	}

	v, err := caseF.Judge(qGate, nil, testTime, delta)
	require.NoError(t, err)
	assert.Equal(t, "fail", v.Outcome().String())

	rulings := v.Rulings()
	require.Len(t, rulings, 2)
	assert.False(t, rulings[0].Passed())
	assert.Contains(t, rulings[0].Detail(), "resolved 1")
}

func TestCaseFileJudgeNilDeltaBackwardsCompatible(t *testing.T) {
	caseF := mustNewCaseFile(t)
	evid := mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, nil)
	require.NoError(t, caseF.AddEvidence(evid, testTime))

	rule, err := qualitygate.NewRule(
		evidence.SubtypeCodeQuality,
		qualitygate.NewZeroTolerance(),
		qualitygate.WithMinResolved(5),
	)
	require.NoError(t, err)
	qGate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)

	v, err := caseF.Judge(qGate, nil, testTime, nil)
	require.NoError(t, err)
	assert.Equal(t, "pass", v.Outcome().String(), "no delta input means delta conditions are skipped")
}

func mustMaxViolations(t *testing.T, max int) qualitygate.MaxViolations {
	t.Helper()
	s, err := qualitygate.NewMaxViolations(max)
	require.NoError(t, err)
	return s
}

func mustCoverageEvidence(t *testing.T, totalLines, coveredLines int, collectedAt time.Time) evidence.Evidence {
	t.Helper()
	content, err := coverage.NewContent(totalLines, coveredLines, nil)
	require.NoError(t, err)
	evid, err := evidence.NewEvidence(evidence.SubtypeCoverage, "test", content, collectedAt)
	require.NoError(t, err)
	return evid
}

func TestCaseFileRecordVerdict(t *testing.T) {
	caseF := mustNewCaseFile(t)

	rulings := []verdict.Ruling{
		mustRuling(t, evidence.SubtypeCodeQuality, true, "0 errors"),
	}
	v, err := verdict.ReconstituteResult("pass", rulings, testTime)
	require.NoError(t, err)

	err = caseF.RecordVerdict(v)

	require.NoError(t, err)
	stored, ok := caseF.Verdict()
	require.True(t, ok)
	assert.Equal(t, "pass", stored.Outcome().String())
	assert.Len(t, stored.Rulings(), 1)
	assert.True(t, caseF.IsJudged())
}

func TestCaseFileRecordVerdictRejectsWhenAlreadyJudged(t *testing.T) {
	caseF := mustNewCaseFile(t)
	evid := mustCodeQualityEvidenceWithFindings(t, "pmd", testTime, nil)
	require.NoError(t, caseF.AddEvidence(evid, testTime))

	qGate := mustQualityGate(t, qualitygate.NewZeroTolerance())
	_, err := caseF.Judge(qGate, nil, testTime, nil)
	require.NoError(t, err)

	v, err := verdict.ReconstituteResult("fail", nil, testTime)
	require.NoError(t, err)

	err = caseF.RecordVerdict(v)

	assert.ErrorIs(t, err, model.ErrAlreadyJudged)
}

func TestCaseFileRecordVerdictEmitsEvents(t *testing.T) {
	caseF := mustNewCaseFile(t)
	caseF.ClearEvents()

	v, err := verdict.ReconstituteResult("fail", nil, testTime)
	require.NoError(t, err)

	err = caseF.RecordVerdict(v)
	require.NoError(t, err)

	events := caseF.Events()
	require.Len(t, events, 1)
	rendered, ok := events[0].(model.VerdictRendered)
	require.True(t, ok)
	assert.True(t, caseF.ID().Equal(rendered.CaseFileID()))
	assert.True(t, caseF.ProjectID().Equal(rendered.ProjectID()))
	assert.Equal(t, "fail", rendered.Outcome())
}

func TestCaseFileRecordVerdictBlocksAddEvidence(t *testing.T) {
	caseF := mustNewCaseFile(t)

	v, err := verdict.ReconstituteResult("pass", nil, testTime)
	require.NoError(t, err)
	require.NoError(t, caseF.RecordVerdict(v))

	evid := mustCodeQualityEvidence(t, "pmd", testTime)
	err = caseF.AddEvidence(evid, testTime)

	assert.ErrorIs(t, err, model.ErrAlreadyJudged)
}

func TestCaseFileRecordVerdictTwiceRejected(t *testing.T) {
	caseF := mustNewCaseFile(t)

	v, err := verdict.ReconstituteResult("pass", nil, testTime)
	require.NoError(t, err)
	require.NoError(t, caseF.RecordVerdict(v))

	err = caseF.RecordVerdict(v)

	assert.ErrorIs(t, err, model.ErrAlreadyJudged)
}

func mustRuling(_ *testing.T, subtype evidence.Subtype, passed bool, detail string) verdict.Ruling {
	return verdict.NewRuling(subtype, passed, detail)
}

func TestCaseFileEvidencesDefensiveCopy(t *testing.T) {
	caseF := mustNewCaseFile(t)
	evid := mustCodeQualityEvidence(t, "pmd", testTime)
	require.NoError(t, caseF.AddEvidence(evid, testTime))

	returned := caseF.Evidences()
	returned[0] = mustCodeQualityEvidence(t, "spotbugs", testTime)

	assert.True(t, evid.ID().Equal(caseF.Evidences()[0].ID()))
}

func mustNewCaseFile(t *testing.T) *model.CaseFile {
	t.Helper()
	caseF, err := model.NewCaseFile(testTenant, mustProjectID(t), "abc123", "main", testTime, testTime)
	require.NoError(t, err)
	return &caseF
}

func mustCodeQualityEvidence(t *testing.T, source string, collectedAt time.Time) evidence.Evidence {
	t.Helper()
	return mustCodeQualityEvidenceWithFindings(t, source, collectedAt, nil)
}

func mustCodeQualityEvidenceWithFindings(t *testing.T, source string, collectedAt time.Time, findings []finding.Finding) evidence.Evidence {
	t.Helper()
	findCont, err := finding.NewContent(evidence.SubtypeCodeQuality, findings)
	require.NoError(t, err)
	evid, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, source, findCont, collectedAt)
	require.NoError(t, err)
	return evid
}

func mustQualityGate(t *testing.T, strategy qualitygate.Strategy) qualitygate.Gate {
	t.Helper()
	rule, err := qualitygate.NewRule(
		evidence.SubtypeCodeQuality,
		strategy,
	)
	require.NoError(t, err)
	qGate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)
	return qGate
}

func verdictPtr(v verdict.Result) *verdict.Result {
	return &v
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

func TestNewCaseFileRejectsZeroTenant(t *testing.T) {
	_, err := model.NewCaseFile(tenant.TenantID{}, mustProjectID(t), "abc123", "main", testTime, testTime)
	require.Error(t, err)
	assert.ErrorIs(t, err, model.ErrInvalidCaseFile)
}
