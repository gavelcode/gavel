package gavelconfig

import (
	gavelspacemodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type WorkspaceFinder struct{}

func NewWorkspaceFinder() *WorkspaceFinder {
	return &WorkspaceFinder{}
}

func (f *WorkspaceFinder) LoadFromConfig(configPath string) (gavelspacemodel.Gavelspace, []projectmodel.Project, error) {
	cfg, err := ParseFile(configPath)
	if err != nil {
		return gavelspacemodel.Gavelspace{}, nil, err
	}
	return cfg.Gavelspace(), cfg.Projects(), nil
}
