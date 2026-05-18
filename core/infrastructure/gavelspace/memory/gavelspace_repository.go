package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/usegavel/gavel/core/domain/gavelspace/model"
	"github.com/usegavel/gavel/core/domain/gavelspace/service"
	"github.com/usegavel/gavel/core/domain/shared/failure"
)

var _ service.GavelspaceRepository = (*GavelspaceRepository)(nil)

var ErrGavelspaceNotFound = failure.New("gavelspace not found", failure.NotFound)

type GavelspaceRepository struct {
	mu     sync.RWMutex
	byName map[string]model.Gavelspace
}

func NewGavelspaceRepository() *GavelspaceRepository {
	return &GavelspaceRepository{
		byName: make(map[string]model.Gavelspace),
	}
}

func (r *GavelspaceRepository) Save(_ context.Context, gs model.Gavelspace) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byName[gs.ID().String()] = gs
	return nil
}

func (r *GavelspaceRepository) FindByName(_ context.Context, name model.GavelspaceID) (model.Gavelspace, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	gs, ok := r.byName[name.String()]
	if !ok {
		return model.Gavelspace{}, fmt.Errorf("%w: %s", ErrGavelspaceNotFound, name.String())
	}
	return gs, nil
}
