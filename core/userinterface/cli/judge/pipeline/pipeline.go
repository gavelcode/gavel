package pipeline

import (
	"context"
	"time"

	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
)

func RunProject(
	ctx context.Context,
	deps Deps,
	workspace string,
	collected collectevidence.Result,
	tenantID, projectID, projectName, commitSHA, branch string,
	startedAt time.Time,
	opts Options,
) (Result, error) {
	if deps.ServerClient != nil {
		return RunServer(ctx, deps, workspace, collected, tenantID, projectID, projectName, commitSHA, branch, startedAt, opts)
	}
	return RunLocal(ctx, deps, workspace, collected, tenantID, projectID, projectName, commitSHA, branch, startedAt, opts)
}
