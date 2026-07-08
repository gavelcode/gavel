package ingestevidence

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
)

var testTenant = tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))

type stubCaseFileRepo struct {
	cf      casefile.CaseFile
	findErr error
	saveErr error
}

func (s *stubCaseFileRepo) Save(_ context.Context, _ casefile.CaseFile) error { return s.saveErr }
func (s *stubCaseFileRepo) FindByID(_ context.Context, _ tenant.TenantID, _ casefile.CaseFileID) (casefile.CaseFile, error) {
	if s.findErr != nil {
		return casefile.CaseFile{}, s.findErr
	}
	return s.cf, nil
}
func (s *stubCaseFileRepo) FindLatestByBranch(_ context.Context, _ projectmodel.ProjectID, _ string) (casefile.CaseFile, error) {
	return casefile.CaseFile{}, nil
}
func (s *stubCaseFileRepo) FindByProject(_ context.Context, _ projectmodel.ProjectID) ([]casefile.CaseFile, error) {
	return nil, nil
}
func (s *stubCaseFileRepo) FindFingerprintIDsByBranch(_ context.Context, _ projectmodel.ProjectID, _ string) ([]finding.FingerprintID, error) {
	return nil, nil
}

func TestNewHandlerPanicsOnNilRepo(t *testing.T) {
	assert.Panics(t, func() { NewHandler(nil) })
}

func TestExecuteParseCaseFileIDError(t *testing.T) {
	handler := &Handler{caseFiles: &stubCaseFileRepo{}}

	cmd := Command{tenantID: testTenant.String(), caseFileID: "not-a-uuid", evidences: []evidencedto.Evidence{{
		Subtype: "code_quality", Source: "test", CollectedAt: time.Now(),
	}}}

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "case file id")
}

func TestExecuteEvidenceToDomainError(t *testing.T) {
	now := time.Now().UTC()
	projectID := projectmodel.NewProjectID(uuid.New())
	cf, err := casefile.NewCaseFile(testTenant, projectID, "sha", "main", now, now)
	require.NoError(t, err)

	handler := &Handler{caseFiles: &stubCaseFileRepo{cf: cf}}

	cmd := Command{tenantID: testTenant.String(), caseFileID: cf.ID().String(), evidences: []evidencedto.Evidence{{
		Subtype: "INVALID", Source: "test", CollectedAt: now,
	}}}

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "evidence[0]")
}

func TestExecuteAddEvidenceErrorAlreadyJudged(t *testing.T) {
	now := time.Now().UTC()
	projectID := projectmodel.NewProjectID(uuid.New())
	caseFile, err := casefile.NewCaseFile(testTenant, projectID, "sha", "main", now, now)
	require.NoError(t, err)

	_, err = caseFile.Judge(qualitygate.Gate{}, nil, now, nil)
	require.NoError(t, err)

	handler := &Handler{caseFiles: &stubCaseFileRepo{cf: caseFile}}

	cmd := Command{tenantID: testTenant.String(), caseFileID: caseFile.ID().String(), evidences: []evidencedto.Evidence{{
		Subtype:     "code_quality",
		Source:      "test",
		CollectedAt: now,
		Findings: []evidencedto.Finding{{
			Tool: "t", RuleID: "r", Severity: "warning",
			FilePath: "f.go", Line: 1, Message: "m", FingerprintID: uuid.NewString(),
		}},
	}}}

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "evidence[0]")
}

func TestExecuteSaveError(t *testing.T) {
	now := time.Now().UTC()
	projectID := projectmodel.NewProjectID(uuid.New())
	cf, err := casefile.NewCaseFile(testTenant, projectID, "sha", "main", now, now)
	require.NoError(t, err)

	handler := &Handler{caseFiles: &stubCaseFileRepo{cf: cf, saveErr: errors.New("disk full")}}

	cmd := Command{tenantID: testTenant.String(), caseFileID: cf.ID().String(), evidences: []evidencedto.Evidence{{
		Subtype:     "code_quality",
		Source:      "test",
		CollectedAt: now,
		Findings: []evidencedto.Finding{{
			Tool: "t", RuleID: "r", Severity: "warning",
			FilePath: "f.go", Line: 1, Message: "m", FingerprintID: uuid.NewString(),
		}},
	}}}

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save case file")
}
