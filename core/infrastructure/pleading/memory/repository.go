package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/usegavel/gavel/core/domain/pleading/model"
	"github.com/usegavel/gavel/core/domain/pleading/service"
	"github.com/usegavel/gavel/core/domain/shared/failure"
)

var _ service.PleadingRepository = (*PleadingRepository)(nil)

var ErrPleadingNotFound = failure.New("pleading not found", failure.NotFound)

type PleadingRepository struct {
	mu   sync.RWMutex
	byID map[string]model.Pleading
}

func NewPleadingRepository() *PleadingRepository {
	return &PleadingRepository{
		byID: make(map[string]model.Pleading),
	}
}

func (r *PleadingRepository) Save(_ context.Context, p model.Pleading) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[p.ID().String()] = p
	return nil
}

func (r *PleadingRepository) FindByID(_ context.Context, id model.PleadingID) (model.Pleading, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.byID[id.String()]
	if !ok {
		return model.Pleading{}, fmt.Errorf("%w: %s", ErrPleadingNotFound, id)
	}
	return p, nil
}
