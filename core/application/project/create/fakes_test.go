package create_test

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
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

func (r *fakeProjectRepo) FindByName(_ context.Context, _ tenant.TenantID, _ string) (projectmodel.Project, error) {
	return projectmodel.Project{}, errNotFound
}

func (r *fakeProjectRepo) FindByKey(_ context.Context, _ tenant.TenantID, _ string) (projectmodel.Project, error) {
	return projectmodel.Project{}, errNotFound
}

func (r *fakeProjectRepo) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.store)
}
