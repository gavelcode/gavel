package file_test

import (
	"context"
	"errors"
	"sync"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/pleading/model"
)

var errNotFound = errors.New("pleading not found")

type fakePleadingRepo struct {
	mu      sync.Mutex
	store   map[string]model.Pleading
	saveErr error
}

func newFakePleadingRepo() *fakePleadingRepo {
	return &fakePleadingRepo{store: make(map[string]model.Pleading)}
}

func (r *fakePleadingRepo) Save(_ context.Context, pleading model.Pleading) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.saveErr != nil {
		return r.saveErr
	}
	r.store[pleading.ID().String()] = pleading
	return nil
}

func (r *fakePleadingRepo) FindByID(_ context.Context, _ tenant.TenantID, pleadingID model.PleadingID) (model.Pleading, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pleading, ok := r.store[pleadingID.String()]
	if !ok {
		return model.Pleading{}, errNotFound
	}
	return pleading, nil
}

func (r *fakePleadingRepo) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.store)
}
