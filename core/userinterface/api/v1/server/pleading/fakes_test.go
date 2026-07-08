package pleading_test

import (
	"context"
	"errors"
	"time"

	pleadingget "github.com/usegavel/gavel/core/application/pleading/get"
	pleadinglist "github.com/usegavel/gavel/core/application/pleading/list"
	"github.com/usegavel/gavel/core/application/project/projectview"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/shared/failure"
	auth "github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

var (
	errFake     = errors.New("fake error")
	errNotFound = failure.New("not found", failure.NotFound)
)

type fakeListFinder struct {
	items []pleadinglist.PleadingSummary
	total int
	err   error
}

func (f *fakeListFinder) ListByProject(_ context.Context, _, _, _ string, _, _ int) ([]pleadinglist.PleadingSummary, int, error) {
	return f.items, f.total, f.err
}

type fakeGetFinder struct {
	detail *pleadingget.PleadingDetail
	err    error
}

func (f *fakeGetFinder) GetByID(_ context.Context, _ string) (*pleadingget.PleadingDetail, error) {
	return f.detail, f.err
}

type fakeGetByKeyFinder struct {
	detail *projectview.ProjectDetail
	err    error
}

func (f *fakeGetByKeyFinder) GetByKey(_ context.Context, _ tenant.TenantID, _ string) (*projectview.ProjectDetail, error) {
	return f.detail, f.err
}

func testPleadingSummary() pleadinglist.PleadingSummary {
	return pleadinglist.PleadingSummary{
		ID:           "44444444-4444-4444-4444-444444444444",
		ProjectID:    "11111111-1111-1111-1111-111111111111",
		Number:       42,
		Title:        "Add login feature",
		Petitioner:   "dev@example.com",
		SourceBranch: "feat/login",
		TargetBranch: "main",
		CommitSHA:    "abc123",
		Status:       "open",
		CreatedAt:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}
}

func testPleadingDetail() *pleadingget.PleadingDetail {
	return &pleadingget.PleadingDetail{
		ID:           "44444444-4444-4444-4444-444444444444",
		ProjectID:    "11111111-1111-1111-1111-111111111111",
		Number:       42,
		Title:        "Add login feature",
		Petitioner:   "dev@example.com",
		SourceBranch: "feat/login",
		TargetBranch: "main",
		CommitSHA:    "abc123",
		Status:       "open",
		CreatedAt:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}
}

func testProjectDetail() *projectview.ProjectDetail {
	return &projectview.ProjectDetail{
		ID:   "11111111-1111-1111-1111-111111111111",
		Key:  "core",
		Name: "Core",
	}
}

const testTenant = "22222222-2222-2222-2222-222222222222"

func authContext() context.Context {
	return auth.WithPrincipal(context.Background(), &auth.Principal{TenantID: testTenant})
}

func testPleadingSummaryWithGate() pleadinglist.PleadingSummary {
	summary := testPleadingSummary()
	summary.GateResult = &pleadinglist.GateResult{
		Passed: true,
		Conditions: []pleadinglist.GateCondition{
			{Label: "coverage", Operator: ">=", Value: "80", Threshold: "70", Passed: true},
		},
	}
	return summary
}

func testPleadingDetailWithGate() *pleadingget.PleadingDetail {
	detail := testPleadingDetail()
	detail.GateResult = &pleadingget.GateResult{
		Passed: true,
		Conditions: []pleadingget.GateCondition{
			{Label: "coverage", Operator: ">=", Value: "80", Threshold: "70", Passed: true},
		},
	}
	return detail
}
