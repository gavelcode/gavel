package createcasefile

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type stubCaseFileRepo struct {
	saveErr error
}

func (s *stubCaseFileRepo) Save(_ context.Context, _ casefile.CaseFile) error { return s.saveErr }
func (s *stubCaseFileRepo) FindByID(_ context.Context, _ casefile.CaseFileID) (casefile.CaseFile, error) {
	return casefile.CaseFile{}, nil
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

type stubProjectRepo struct {
	project projectmodel.Project
	findErr error
}

func (s *stubProjectRepo) Save(_ context.Context, _ projectmodel.Project) error { return nil }
func (s *stubProjectRepo) FindByID(_ context.Context, _ projectmodel.ProjectID) (projectmodel.Project, error) {
	if s.findErr != nil {
		return projectmodel.Project{}, s.findErr
	}
	return s.project, nil
}
func (s *stubProjectRepo) FindByName(_ context.Context, _ string) (projectmodel.Project, error) {
	return projectmodel.Project{}, nil
}
func (s *stubProjectRepo) FindByKey(_ context.Context, _ string) (projectmodel.Project, error) {
	return projectmodel.Project{}, nil
}

func TestExecuteNewCaseFileDomainError(t *testing.T) {
	project, err := projectmodel.NewProject("test", "test", "//...")
	require.NoError(t, err)

	handler := &Handler{
		caseFiles: &stubCaseFileRepo{},
		projects:  &stubProjectRepo{project: project},
	}

	cmd := Command{
		projectID: uuid.NewString(),
		commitSHA: "abc",
		branch:    "main",
		startedAt: time.Time{},
	}

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "new case file")
}

func TestExecuteSaveError(t *testing.T) {
	project, err := projectmodel.NewProject("test", "test", "//...")
	require.NoError(t, err)

	handler := &Handler{
		caseFiles: &stubCaseFileRepo{saveErr: errors.New("disk full")},
		projects:  &stubProjectRepo{project: project},
	}

	cmd := Command{
		projectID: project.ID().String(),
		commitSHA: "abc",
		branch:    "main",
		startedAt: time.Now().UTC(),
	}

	_, err = handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save case file")
}
