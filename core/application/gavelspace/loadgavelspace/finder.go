package loadgavelspace

import (
	gavelspacemodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type Finder interface {
	LoadFromConfig(configPath string) (gavelspacemodel.Gavelspace, []projectmodel.Project, error)
}
