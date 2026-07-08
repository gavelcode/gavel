package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/service"
	"github.com/usegavel/gavel/core/domain/shared/failure"
)

var _ service.ProjectRepository = (*ProjectRepository)(nil)

var ErrProjectNotFound = failure.New("project not found", failure.NotFound)

type tenantProjects struct {
	byID   map[string]model.Project
	byName map[string]model.Project
	byKey  map[string]model.Project
}

func newTenantProjects() *tenantProjects {
	return &tenantProjects{
		byID:   make(map[string]model.Project),
		byName: make(map[string]model.Project),
		byKey:  make(map[string]model.Project),
	}
}

type ProjectRepository struct {
	mu            sync.RWMutex
	byTenant      map[string]*tenantProjects
	baselineStore *BaselineStore
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{
		byTenant: make(map[string]*tenantProjects),
	}
}

func NewProjectRepositoryWithBaseline(store *BaselineStore) *ProjectRepository {
	return &ProjectRepository{
		byTenant:      make(map[string]*tenantProjects),
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

	tenantKey := project.TenantID().String()
	scoped := r.byTenant[tenantKey]
	if scoped == nil {
		scoped = newTenantProjects()
		r.byTenant[tenantKey] = scoped
	}
	scoped.byID[project.ID().String()] = project
	scoped.byName[project.Name()] = project
	scoped.byKey[project.Key()] = project
	return nil
}

func (r *ProjectRepository) FindByID(_ context.Context, tenantID tenant.TenantID, projectID model.ProjectID) (model.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if scoped := r.byTenant[tenantID.String()]; scoped != nil {
		if p, ok := scoped.byID[projectID.String()]; ok {
			return p, nil
		}
	}
	return model.Project{}, fmt.Errorf("%w: %s", ErrProjectNotFound, projectID)
}

func (r *ProjectRepository) FindByName(_ context.Context, tenantID tenant.TenantID, name string) (model.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if scoped := r.byTenant[tenantID.String()]; scoped != nil {
		if p, ok := scoped.byName[name]; ok {
			return p, nil
		}
	}
	return model.Project{}, fmt.Errorf("%w: %s", ErrProjectNotFound, name)
}

func (r *ProjectRepository) FindByKey(_ context.Context, tenantID tenant.TenantID, key string) (model.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if scoped := r.byTenant[tenantID.String()]; scoped != nil {
		if p, ok := scoped.byKey[key]; ok {
			return p, nil
		}
	}
	return model.Project{}, fmt.Errorf("%w: %s", ErrProjectNotFound, key)
}
