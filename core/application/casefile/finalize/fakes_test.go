package finalize_test

import (
	"context"
	"errors"
	"sync"

	"github.com/usegavel/gavel/core/application/casefile/finalize"
	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

var errNotFound = errors.New("not found")

type fakeCaseFileRepo struct {
	mu       sync.Mutex
	store    map[string]casefile.CaseFile
	fpStore  map[string][]finding.FingerprintID
	findErr  error
	saveErr  error
	fpErr    error
}

func newFakeCaseFileRepo() *fakeCaseFileRepo {
	return &fakeCaseFileRepo{
		store:   make(map[string]casefile.CaseFile),
		fpStore: make(map[string][]finding.FingerprintID),
	}
}

func (r *fakeCaseFileRepo) Save(_ context.Context, caseFile casefile.CaseFile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.saveErr != nil {
		return r.saveErr
	}
	r.store[caseFile.ID().String()] = caseFile
	return nil
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

type fakeCounterWriter struct {
	lastCaseFileID string
	lastWritten    *finalize.Counters
	err            error
}

func (w *fakeCounterWriter) WriteCounters(_ context.Context, caseFileID string, counters finalize.Counters) error {
	w.lastCaseFileID = caseFileID
	c := counters
	w.lastWritten = &c
	return w.err
}

type fakeProjectRepo struct {
	mu      sync.Mutex
	store   map[string]projectmodel.Project
	saveErr error
	saved   []projectmodel.Project
}

func newFakeProjectRepo() *fakeProjectRepo {
	return &fakeProjectRepo{store: make(map[string]projectmodel.Project)}
}

func (r *fakeProjectRepo) seed(project projectmodel.Project) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[project.ID().String()] = project
}

func (r *fakeProjectRepo) Save(_ context.Context, project projectmodel.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.saveErr != nil {
		return r.saveErr
	}
	r.store[project.ID().String()] = project
	r.saved = append(r.saved, project)
	return nil
}

func (r *fakeProjectRepo) FindByID(_ context.Context, projectID projectmodel.ProjectID) (projectmodel.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	project, ok := r.store[projectID.String()]
	if !ok {
		return projectmodel.Project{}, errNotFound
	}
	return project, nil
}

func (r *fakeProjectRepo) FindByName(_ context.Context, _ string) (projectmodel.Project, error) {
	return projectmodel.Project{}, errNotFound
}

func (r *fakeProjectRepo) FindByKey(_ context.Context, _ string) (projectmodel.Project, error) {
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
