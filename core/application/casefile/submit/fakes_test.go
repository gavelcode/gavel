package submit_test

import (
	"context"
	"errors"
	"sync"

	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

var errNotFound = errors.New("not found")

type fakeCaseFileRepo struct {
	mu    sync.Mutex
	store map[string]casefile.CaseFile
	fpErr error
}

func newFakeCaseFileRepo() *fakeCaseFileRepo {
	return &fakeCaseFileRepo{store: make(map[string]casefile.CaseFile)}
}

func (r *fakeCaseFileRepo) Save(_ context.Context, cf casefile.CaseFile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[cf.ID().String()] = cf
	return nil
}

func (r *fakeCaseFileRepo) FindByID(_ context.Context, _ tenant.TenantID, id casefile.CaseFileID) (casefile.CaseFile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cf, ok := r.store[id.String()]
	if !ok {
		return casefile.CaseFile{}, errNotFound
	}
	return cf, nil
}

func (r *fakeCaseFileRepo) FindLatestByBranch(_ context.Context, _ projectmodel.ProjectID, _ string) (casefile.CaseFile, error) {
	return casefile.CaseFile{}, errNotFound
}

func (r *fakeCaseFileRepo) FindByProject(_ context.Context, _ projectmodel.ProjectID) ([]casefile.CaseFile, error) {
	return nil, nil
}

func (r *fakeCaseFileRepo) FindFingerprintIDsByBranch(_ context.Context, _ projectmodel.ProjectID, _ string) ([]finding.FingerprintID, error) {
	if r.fpErr != nil {
		return nil, r.fpErr
	}
	return nil, nil
}

type fakeProjectRepo struct {
	mu    sync.Mutex
	store map[string]projectmodel.Project
	saved []projectmodel.Project
}

func newFakeProjectRepo() *fakeProjectRepo {
	return &fakeProjectRepo{store: make(map[string]projectmodel.Project)}
}

func (r *fakeProjectRepo) seed(p projectmodel.Project) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[p.ID().String()] = p
}

func (r *fakeProjectRepo) Save(_ context.Context, p projectmodel.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[p.ID().String()] = p
	r.saved = append(r.saved, p)
	return nil
}

func (r *fakeProjectRepo) FindByID(_ context.Context, _ tenant.TenantID, id projectmodel.ProjectID) (projectmodel.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.store[id.String()]
	if !ok {
		return projectmodel.Project{}, errNotFound
	}
	return p, nil
}

func (r *fakeProjectRepo) FindByName(_ context.Context, _ tenant.TenantID, _ string) (projectmodel.Project, error) {
	return projectmodel.Project{}, errNotFound
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
