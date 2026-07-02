package v1

import (
	"github.com/usegavel/gavel/core/userinterface/api/v1/gen"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/casefile"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/gavelspace"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/iam"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/ops"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/pleading"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/project"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/supporting/search"
	"github.com/usegavel/gavel/core/userinterface/api/v1/server/supporting/source"
)

var _ gen.StrictServerInterface = (*Server)(nil)

type (
	CaseFileHandler   = casefile.Handler
	GavelspaceHandler = gavelspace.Handler
	IAMHandler        = iam.Handler
	OpsHandler        = ops.Handler
	PleadingHandler   = pleading.Handler
	ProjectHandler    = project.Handler
	SearchHandler     = search.Handler
	SourceHandler     = source.Handler
)

type Server struct {
	*CaseFileHandler
	*GavelspaceHandler
	*IAMHandler
	*OpsHandler
	*PleadingHandler
	*ProjectHandler
	*SearchHandler
	*SourceHandler
}
