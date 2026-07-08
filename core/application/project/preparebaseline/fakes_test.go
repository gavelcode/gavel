package preparebaseline_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/project/preparebaseline"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

var (
	errNotFound = errors.New("not found")
	testTenant  = tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))
)

type fakeProjectRepo struct {
	mu      sync.Mutex
	store   map[string]projectmodel.Project
	saved   []projectmodel.Project
	saveErr error
}

func newFakeProjectRepo() *fakeProjectRepo {
	return &fakeProjectRepo{store: make(map[string]projectmodel.Project)}
}

func (r *fakeProjectRepo) Save(_ context.Context, project projectmodel.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.saveErr != nil {
		return r.saveErr
	}
	r.store[project.ID().String()] = project
	r.store["name:"+project.Name()] = project
	r.saved = append(r.saved, project)
	return nil
}

func (r *fakeProjectRepo) FindByID(_ context.Context, _ tenant.TenantID, id projectmodel.ProjectID) (projectmodel.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	project, ok := r.store[id.String()]
	if !ok {
		return projectmodel.Project{}, errNotFound
	}
	return project, nil
}

func (r *fakeProjectRepo) FindByName(_ context.Context, _ tenant.TenantID, name string) (projectmodel.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	project, ok := r.store["name:"+name]
	if !ok {
		return projectmodel.Project{}, errNotFound
	}
	return project, nil
}

func (r *fakeProjectRepo) FindByKey(_ context.Context, _ tenant.TenantID, _ string) (projectmodel.Project, error) {
	return projectmodel.Project{}, errNotFound
}

func (r *fakeProjectRepo) lastSaved() projectmodel.Project {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.saved) == 0 {
		return projectmodel.Project{}
	}
	return r.saved[len(r.saved)-1]
}

type fakeCaseFileRepo struct {
	mu        sync.Mutex
	preloaded map[string]int
}

func newFakeCaseFileRepo() *fakeCaseFileRepo {
	return &fakeCaseFileRepo{preloaded: make(map[string]int)}
}

func (r *fakeCaseFileRepo) PreloadFingerprints(projectID projectmodel.ProjectID, branch string, fps []finding.FingerprintID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.preloaded[projectID.String()+"|"+branch] = len(fps)
}

func (r *fakeCaseFileRepo) preloadedCount(projectID projectmodel.ProjectID, branch string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.preloaded[projectID.String()+"|"+branch]
}

type fakeFetcher struct {
	baselines map[string]*preparebaseline.RemoteBaseline
}

func (f *fakeFetcher) FetchBaseline(_ context.Context, projectKey, branch string) (*preparebaseline.RemoteBaseline, error) {
	bl, ok := f.baselines[projectKey+"|"+branch]
	if !ok {
		return nil, errors.New("not found")
	}
	return bl, nil
}

func seedProject(t *testing.T, repo *fakeProjectRepo, name string) projectmodel.Project {
	t.Helper()
	project, err := projectmodel.NewProject(testTenant, name, name, "//"+name+"/...")
	require.NoError(t, err)
	require.NoError(t, repo.Save(context.Background(), project))
	return project
}

func projectIDByName(repo *fakeProjectRepo, name string) projectmodel.ProjectID {
	project, _ := repo.FindByName(context.Background(), testTenant, name)
	return project.ID()
}
