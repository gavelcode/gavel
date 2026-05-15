package preparebaseline

import (
	"context"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	projectservice "github.com/usegavel/gavel/core/domain/project/service"
)

type BaselineFetcher interface {
	FetchBaseline(ctx context.Context, projectKey, branch string) (*RemoteBaseline, error)
}

type RemoteBaseline struct {
	Fingerprints     []string
	ArchViolationIDs []string
	HasPrevious      bool
}

type FingerprintSeeder interface {
	PreloadFingerprints(projectID projectmodel.ProjectID, branch string, fingerprints []finding.FingerprintID)
}

type ProjectRepository = projectservice.ProjectRepository
