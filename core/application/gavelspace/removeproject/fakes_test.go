package removeproject_test

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	gsmodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

var errNotFound = errors.New("not found")

type fakeGavelspaceRepo struct {
	mu      sync.Mutex
	store   map[string]gsmodel.Gavelspace
	findErr error
	saveErr error
}

func newFakeGavelspaceRepo() *fakeGavelspaceRepo {
	return &fakeGavelspaceRepo{store: make(map[string]gsmodel.Gavelspace)}
}

func (r *fakeGavelspaceRepo) seed(gavelspace gsmodel.Gavelspace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[gavelspace.ID().String()] = gavelspace
}

func (r *fakeGavelspaceRepo) FindByName(_ context.Context, _ tenant.TenantID, name gsmodel.GavelspaceID) (gsmodel.Gavelspace, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.findErr != nil {
		return gsmodel.Gavelspace{}, r.findErr
	}
	gavelspace, ok := r.store[name.String()]
	if !ok {
		return gsmodel.Gavelspace{}, errNotFound
	}
	return gavelspace, nil
}

func (r *fakeGavelspaceRepo) Save(_ context.Context, gavelspace gsmodel.Gavelspace) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.saveErr != nil {
		return r.saveErr
	}
	r.store[gavelspace.ID().String()] = gavelspace
	return nil
}

const testTenant = "22222222-2222-2222-2222-222222222222"

var testTenantID = tenant.NewTenantID(uuid.MustParse(testTenant))
