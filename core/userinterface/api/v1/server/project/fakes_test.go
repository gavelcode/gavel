package project_test

import (
	"context"
	"errors"
	"time"

	projectlist "github.com/usegavel/gavel/core/application/project/list"
	"github.com/usegavel/gavel/core/application/project/projectview"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/shared/failure"
	memproject "github.com/usegavel/gavel/core/infrastructure/project/memory"
)

var (
	errFake     = errors.New("fake error")
	errNotFound = failure.New("not found", failure.NotFound)
)

type fakeListFinder struct {
	items []projectlist.ProjectSummary
	total int
	err   error
}

func (f *fakeListFinder) List(_ context.Context, _, _ int) ([]projectlist.ProjectSummary, int, error) {
	return f.items, f.total, f.err
}

type fakeGetByKeyFinder struct {
	detail *projectview.ProjectDetail
	err    error
}

func (f *fakeGetByKeyFinder) GetByKey(_ context.Context, _ string) (*projectview.ProjectDetail, error) {
	return f.detail, f.err
}

func testProjectDetail() *projectview.ProjectDetail {
	return &projectview.ProjectDetail{
		ID:            "00000000-0000-0000-0000-000000000001",
		Key:           "core",
		Name:          "Core Module",
		DefaultBranch: "main",
		LatestVerdict: "pass",
		TotalFindings: 42,
		CreatedAt:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		TargetPattern: "//core/...",
		Languages:     []string{"go"},
		QualityGateRules: []projectview.QualityGateRuleView{
			{Subtype: "code_quality", StrategyType: "zero_tolerance"},
		},
		SeverityCounts: map[string]int{"error": 10, "warning": 32},
	}
}

func testProjectSummary() projectlist.ProjectSummary {
	return projectlist.ProjectSummary{
		ID:            "00000000-0000-0000-0000-000000000001",
		Key:           "core",
		Name:          "Core Module",
		DefaultBranch: "main",
		LatestVerdict: "pass",
		TotalFindings: 42,
		CreatedAt:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func detailWithID(id string) *projectview.ProjectDetail {
	detail := testProjectDetail()
	detail.ID = id
	return detail
}

type conflictRepo struct {
	inner *memproject.ProjectRepository
}

func newConflictRepo() *conflictRepo {
	return &conflictRepo{inner: memproject.NewProjectRepository()}
}

func (r *conflictRepo) Save(_ context.Context, _ projectmodel.Project) error {
	return failure.New("project key already exists", failure.Conflict)
}

func (r *conflictRepo) FindByID(ctx context.Context, id projectmodel.ProjectID) (projectmodel.Project, error) {
	return r.inner.FindByID(ctx, id)
}

func (r *conflictRepo) FindByName(ctx context.Context, name string) (projectmodel.Project, error) {
	return r.inner.FindByName(ctx, name)
}

func (r *conflictRepo) FindByKey(ctx context.Context, key string) (projectmodel.Project, error) {
	return r.inner.FindByKey(ctx, key)
}
