package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/service"
	"github.com/usegavel/gavel/core/domain/shared/failure"
)

var _ service.ProjectRepository = (*ProjectRepository)(nil)

var ErrProjectNotFound = failure.New("project not found", failure.NotFound)

type ProjectRepository struct {
	mu            sync.RWMutex
	byID          map[string]model.Project
	byName        map[string]model.Project
	byKey         map[string]model.Project
	baselineStore *BaselineStore
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{
		byID:   make(map[string]model.Project),
		byName: make(map[string]model.Project),
		byKey:  make(map[string]model.Project),
	}
}

func NewProjectRepositoryWithBaseline(store *BaselineStore) *ProjectRepository {
	return &ProjectRepository{
		byID:          make(map[string]model.Project),
		byName:        make(map[string]model.Project),
		byKey:         make(map[string]model.Project),
		baselineStore: store,
	}
}

func (r *ProjectRepository) SetBaselineStore(store *BaselineStore) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.baselineStore = store
}

func (r *ProjectRepository) Save(_ context.Context, project model.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.baselineStore != nil {
		if baselines := project.Baselines(); len(baselines) > 0 {
			if bl, ok := baselines[project.DefaultBranch()]; ok {
				if err := r.baselineStore.Save(project.Name(), bl); err != nil {
					return fmt.Errorf("save baselines: %w", err)
				}
			}
		} else {
			bl := r.baselineStore.Load(project.Name())
			if bl.HasPrevious() {
				project.UpdateBaseline(project.DefaultBranch(), bl.Fingerprints(), bl.ArchIDs(), bl.CoveragePercent(), bl.FileCoverage())
			}
		}
	}

	r.byID[project.ID().String()] = project
	r.byName[project.Name()] = project
	r.byKey[project.Key()] = project
	return nil
}

func (r *ProjectRepository) FindByID(_ context.Context, projectID model.ProjectID) (model.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.byID[projectID.String()]
	if !ok {
		return model.Project{}, fmt.Errorf("%w: %s", ErrProjectNotFound, projectID)
	}
	return p, nil
}

func (r *ProjectRepository) FindByName(_ context.Context, name string) (model.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.byName[name]
	if !ok {
		return model.Project{}, fmt.Errorf("%w: %s", ErrProjectNotFound, name)
	}
	return p, nil
}

func (r *ProjectRepository) FindByKey(_ context.Context, key string) (model.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.byKey[key]
	if !ok {
		return model.Project{}, fmt.Errorf("%w: %s", ErrProjectNotFound, key)
	}
	return p, nil
}
