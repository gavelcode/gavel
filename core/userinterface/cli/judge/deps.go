package judge

import (
	"context"
	"log/slog"

	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	ingestcov "github.com/usegavel/gavel/core/application/casefile/ingestcoverage"
	ingestfind "github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	"github.com/usegavel/gavel/core/application/casefile/submit"
	"github.com/usegavel/gavel/core/application/gavelspace/loadgavelspace"
	"github.com/usegavel/gavel/core/application/project/preparebaseline"
	apiclient "github.com/usegavel/gavel/core/userinterface/api/v1/client"
)

type SourceContext interface {
	CommitSHA(ctx context.Context) (string, error)
	Branch(ctx context.Context) (string, error)
	ChangedLines(ctx context.Context, workspace, baseRef string) (map[string][]int, error)
}

type TargetQuery interface {
	QueryTargetsOfKind(ctx context.Context, workspace, pattern string, kinds []string) ([]string, error)
}

type TargetResolver interface {
	FindAffectedTargets(ctx context.Context, workspace string, changedFiles []string, scope string) ([]string, error)
	FindOwnerTarget(ctx context.Context, workspace, file string) (string, error)
}

type StructureVerifier interface {
	VerifyStructure(workspace string) ([]string, error)
}

type WorkspaceResolver func() (string, error)

type deps struct {
	findings         *ingestfind.Handler
	coverage         *ingestcov.Handler
	submitH          *submit.Handler
	collectEvH       *collectevidence.Handler
	loadWorkspace    *loadgavelspace.Handler
	projectRepo      preparebaseline.ProjectRepository
	fpSeeder         preparebaseline.FingerprintSeeder
	resolveWorkspace WorkspaceResolver
	source           SourceContext
	validate         StructureVerifier
	log              *slog.Logger
	serverClient     *apiclient.Client
	targetQuery      TargetQuery
	targetResolver   TargetResolver
}
