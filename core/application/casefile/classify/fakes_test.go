package classify_test

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
	mu              sync.Mutex
	fingerprints    map[string][]finding.FingerprintID
	fingerprintsErr error
}

func newFakeCaseFileRepo() *fakeCaseFileRepo {
	return &fakeCaseFileRepo{fingerprints: make(map[string][]finding.FingerprintID)}
}

func (r *fakeCaseFileRepo) seedFingerprints(projectID projectmodel.ProjectID, branch string, fps []finding.FingerprintID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fingerprints[projectID.String()+"|"+branch] = fps
}

func (r *fakeCaseFileRepo) FindFingerprintIDsByBranch(_ context.Context, projectID projectmodel.ProjectID, branch string) ([]finding.FingerprintID, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.fingerprintsErr != nil {
		return nil, r.fingerprintsErr
	}
	return r.fingerprints[projectID.String()+"|"+branch], nil
}

func (r *fakeCaseFileRepo) FindByID(_ context.Context, _ tenant.TenantID, _ casefile.CaseFileID) (casefile.CaseFile, error) {
	return casefile.CaseFile{}, errNotFound
}

func (r *fakeCaseFileRepo) FindLatestByBranch(_ context.Context, _ projectmodel.ProjectID, _ string) (casefile.CaseFile, error) {
	return casefile.CaseFile{}, errNotFound
}

func (r *fakeCaseFileRepo) Save(_ context.Context, _ casefile.CaseFile) error {
	return nil
}

func (r *fakeCaseFileRepo) FindByProject(_ context.Context, _ projectmodel.ProjectID) ([]casefile.CaseFile, error) {
	return nil, nil
}
