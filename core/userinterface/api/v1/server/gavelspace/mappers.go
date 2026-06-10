package gavelspace

import (
	gsget "github.com/usegavel/gavel/core/application/gavelspace/get"
	gslist "github.com/usegavel/gavel/core/application/gavelspace/list"
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx"
)

func gavelspaceFromSummary(s gslist.GavelspaceSummary) gen.Gavelspace {
	return gen.Gavelspace{
		Name:         s.Name,
		ProjectCount: int32(s.ProjectCount),
		CreatedAt:    s.CreatedAt,
	}
}

func projectRefFromView(p gsget.ProjectRefView) gen.GavelspaceProject {
	return gen.GavelspaceProject{
		Id:            httpx.ParseUUIDOrZero(p.ID),
		Key:           p.Key,
		Name:          p.Name,
		LatestVerdict: p.LatestVerdict,
	}
}
