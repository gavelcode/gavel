package service

import (
	"context"

	"github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

type CaseFileRepository interface {
	Save(ctx context.Context, caseFile model.CaseFile) error
	FindByID(ctx context.Context, tenantID tenant.TenantID, id model.CaseFileID) (model.CaseFile, error)
	FindByProject(ctx context.Context, projectID projectmodel.ProjectID) ([]model.CaseFile, error)
	FindLatestByBranch(ctx context.Context, projectID projectmodel.ProjectID, branch string) (model.CaseFile, error)
	FindFingerprintIDsByBranch(ctx context.Context, projectID projectmodel.ProjectID, branch string) ([]finding.FingerprintID, error)
}
