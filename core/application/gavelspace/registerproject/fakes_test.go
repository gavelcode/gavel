package registerproject_test

import (
	"context"
	"fmt"

	gsmodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
)

type fakeGavelspaceRepo struct {
	gavelspaces map[string]gsmodel.Gavelspace
	saveErr     error
}

func newFakeGavelspaceRepo() *fakeGavelspaceRepo {
	return &fakeGavelspaceRepo{gavelspaces: map[string]gsmodel.Gavelspace{}}
}

func (r *fakeGavelspaceRepo) seed(gs gsmodel.Gavelspace) {
	r.gavelspaces[gs.ID().String()] = gs
}

func (r *fakeGavelspaceRepo) FindByName(_ context.Context, name gsmodel.GavelspaceID) (gsmodel.Gavelspace, error) {
	gs, ok := r.gavelspaces[name.String()]
	if !ok {
		return gsmodel.Gavelspace{}, fmt.Errorf("gavelspace not found: %s", name)
	}
	return gs, nil
}

func (r *fakeGavelspaceRepo) Save(_ context.Context, gs gsmodel.Gavelspace) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.gavelspaces[gs.ID().String()] = gs
	return nil
}
