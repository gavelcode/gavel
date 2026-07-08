package judge_test

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
	mu        sync.Mutex
	store     map[string]casefile.CaseFile
	findErr   error
	saveErr   error
	saveCalls int
}

func newFakeCaseFileRepo() *fakeCaseFileRepo {
	return &fakeCaseFileRepo{store: make(map[string]casefile.CaseFile)}
}

func (r *fakeCaseFileRepo) seed(caseFile casefile.CaseFile) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[caseFile.ID().String()] = caseFile
}

func (r *fakeCaseFileRepo) FindByID(_ context.Context, caseFileID casefile.CaseFileID) (casefile.CaseFile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.findErr != nil {
		return casefile.CaseFile{}, r.findErr
	}
	caseFile, ok := r.store[caseFileID.String()]
	if !ok {
		return casefile.CaseFile{}, errNotFound
	}
	return caseFile, nil
}

func (r *fakeCaseFileRepo) Save(_ context.Context, caseFile casefile.CaseFile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.saveCalls++
	if r.saveErr != nil {
		return r.saveErr
	}
	r.store[caseFile.ID().String()] = caseFile
	return nil
}

func (r *fakeCaseFileRepo) FindLatestByBranch(_ context.Context, _ projectmodel.ProjectID, _ string) (casefile.CaseFile, error) {
	return casefile.CaseFile{}, errNotFound
}

func (r *fakeCaseFileRepo) FindFingerprintIDsByBranch(_ context.Context, _ projectmodel.ProjectID, _ string) ([]finding.FingerprintID, error) {
	return nil, nil
}

func (r *fakeCaseFileRepo) FindByProject(_ context.Context, _ projectmodel.ProjectID) ([]casefile.CaseFile, error) {
	return nil, nil
}

type fakeProjectRepo struct {
	mu      sync.Mutex
	store   map[string]projectmodel.Project
	findErr error
}

func newFakeProjectRepo() *fakeProjectRepo {
	return &fakeProjectRepo{store: make(map[string]projectmodel.Project)}
}

func (r *fakeProjectRepo) seed(project projectmodel.Project) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[project.ID().String()] = project
}

func (r *fakeProjectRepo) FindByID(_ context.Context, _ tenant.TenantID, projectID projectmodel.ProjectID) (projectmodel.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.findErr != nil {
		return projectmodel.Project{}, r.findErr
	}
	project, ok := r.store[projectID.String()]
	if !ok {
		return projectmodel.Project{}, errNotFound
	}
	return project, nil
}

func (r *fakeProjectRepo) Save(_ context.Context, _ projectmodel.Project) error {
	return nil
}

func (r *fakeProjectRepo) FindByName(_ context.Context, _ tenant.TenantID, _ string) (projectmodel.Project, error) {
	return projectmodel.Project{}, errNotFound
}

func (r *fakeProjectRepo) FindByKey(_ context.Context, _ tenant.TenantID, _ string) (projectmodel.Project, error) {
	return projectmodel.Project{}, errNotFound
}
