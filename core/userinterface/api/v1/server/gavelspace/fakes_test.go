package gavelspace_test

import (
	"context"
	"errors"
	"time"

	gsget "github.com/usegavel/gavel/core/application/gavelspace/get"
	gslist "github.com/usegavel/gavel/core/application/gavelspace/list"
	gsmodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	gsservice "github.com/usegavel/gavel/core/domain/gavelspace/service"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/shared/failure"
	auth "github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

var (
	errFake     = errors.New("fake error")
	errNotFound = failure.New("not found", failure.NotFound)
)

type fakeListFinder struct {
	items []gslist.GavelspaceSummary
	total int
	err   error
}

func (f *fakeListFinder) List(_ context.Context, _ tenant.TenantID, _, _ int) ([]gslist.GavelspaceSummary, int, error) {
	return f.items, f.total, f.err
}

type fakeGetFinder struct {
	detail *gsget.GavelspaceDetail
	err    error
}

func (f *fakeGetFinder) GetByName(_ context.Context, _ tenant.TenantID, _ string) (*gsget.GavelspaceDetail, error) {
	return f.detail, f.err
}

type conflictRepo struct {
	inner gsservice.GavelspaceRepository
}

func (r *conflictRepo) Save(_ context.Context, _ gsmodel.Gavelspace) error {
	return failure.New("gavelspace already exists", failure.Conflict)
}

func (r *conflictRepo) FindByName(ctx context.Context, tenantID tenant.TenantID, name gsmodel.GavelspaceID) (gsmodel.Gavelspace, error) {
	return r.inner.FindByName(ctx, tenantID, name)
}

func testSummary() gslist.GavelspaceSummary {
	return gslist.GavelspaceSummary{
		Name:         "gavel",
		ProjectCount: 3,
		CreatedAt:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func testDetail() *gsget.GavelspaceDetail {
	return &gsget.GavelspaceDetail{
		Name: "gavel",
		Projects: []gsget.ProjectRefView{
			{ID: "00000000-0000-0000-0000-000000000001", Key: "core", Name: "Core", LatestVerdict: "pass"},
		},
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

const testTenant = "22222222-2222-2222-2222-222222222222"

func authContext() context.Context {
	return auth.WithPrincipal(context.Background(), &auth.Principal{TenantID: testTenant})
}
