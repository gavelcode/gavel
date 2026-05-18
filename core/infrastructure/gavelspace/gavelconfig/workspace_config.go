package gavelconfig

import (
	gavelspacemodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type WorkspaceConfig struct {
	gavelspace      gavelspacemodel.Gavelspace
	projects        []projectmodel.Project
	coverageOptions map[string]CoverageOptions
	server          ServerConfig
	findingsSource  string
}

func (w WorkspaceConfig) Gavelspace() gavelspacemodel.Gavelspace { return w.gavelspace }

func (w WorkspaceConfig) Projects() []projectmodel.Project {
	copied := make([]projectmodel.Project, len(w.projects))
	copy(copied, w.projects)
	return copied
}

func (w WorkspaceConfig) CoverageOptionsForProject(name string) CoverageOptions {
	return w.coverageOptions[name]
}

func (w WorkspaceConfig) Server() ServerConfig   { return w.server }
func (w WorkspaceConfig) FindingsSource() string { return w.findingsSource }
