package getbaseline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/project/projectview"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type fakeProjectFinder struct {
	detail *projectview.ProjectDetail
	err    error
}

func (f *fakeProjectFinder) GetByKey(_ context.Context, _ string) (*projectview.ProjectDetail, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.detail, nil
}

type fakeProjectRepo struct {
	project projectmodel.Project
	err     error
}

func (f *fakeProjectRepo) Save(_ context.Context, _ projectmodel.Project) error {
	return nil
}

func (f *fakeProjectRepo) FindByID(_ context.Context, _ projectmodel.ProjectID) (projectmodel.Project, error) {
	if f.err != nil {
		return projectmodel.Project{}, f.err
	}
	return f.project, nil
}

func (f *fakeProjectRepo) FindByName(_ context.Context, _ string) (projectmodel.Project, error) {
	return f.project, f.err
}

func (f *fakeProjectRepo) FindByKey(_ context.Context, _ string) (projectmodel.Project, error) {
	return f.project, f.err
}

func seedProjectWithBaseline(t *testing.T, branch string, fingerprints, archIDs []string) projectmodel.Project {
	t.Helper()
	p, err := projectmodel.NewProject("core", "Core", "//core/...")
	require.NoError(t, err)
	p.UpdateBaseline(branch, fingerprints, archIDs, nil, nil)
	return p
}

func seedProjectWithoutBaseline(t *testing.T) projectmodel.Project {
	t.Helper()
	p, err := projectmodel.NewProject("core", "Core", "//core/...")
	require.NoError(t, err)
	return p
}

func projectDetail(key string, p projectmodel.Project) *projectview.ProjectDetail {
	return &projectview.ProjectDetail{
		ID:            p.ID().String(),
		Key:           key,
		DefaultBranch: "main",
	}
}

func projectDetailWithBranch(key, branch string, p projectmodel.Project) *projectview.ProjectDetail {
	return &projectview.ProjectDetail{
		ID:            p.ID().String(),
		Key:           key,
		DefaultBranch: branch,
	}
}
