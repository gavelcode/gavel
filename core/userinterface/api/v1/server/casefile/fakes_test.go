package casefile_test

import (
	"context"
	"errors"
	"time"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	casefileget "github.com/usegavel/gavel/core/application/casefile/get"
	casefilelist "github.com/usegavel/gavel/core/application/casefile/list"
	findinglist "github.com/usegavel/gavel/core/application/casefile/listfindings"
	"github.com/usegavel/gavel/core/application/project/projectview"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/shared/failure"
)

var (
	errFake     = errors.New("fake error")
	errNotFound = failure.New("not found", failure.NotFound)
)

type fakeListFinder struct {
	items []casefilelist.CaseFileSummary
	total int
	err   error
}

func (f *fakeListFinder) ListByProject(_ context.Context, _ tenant.TenantID, _, _ string, _, _ int) ([]casefilelist.CaseFileSummary, int, error) {
	return f.items, f.total, f.err
}

type fakeGetFinder struct {
	detail *casefileget.CaseFileDetail
	err    error
}

func (f *fakeGetFinder) GetByID(_ context.Context, _ tenant.TenantID, _ string) (*casefileget.CaseFileDetail, error) {
	return f.detail, f.err
}

type fakeFindingFinder struct {
	items []findinglist.FindingView
	total int
	err   error
}

func (f *fakeFindingFinder) List(_ context.Context, _ tenant.TenantID, _ findinglist.Filters, _, _ int) ([]findinglist.FindingView, int, error) {
	return f.items, f.total, f.err
}

type fakeGetByKeyFinder struct {
	detail *projectview.ProjectDetail
	err    error
}

func (f *fakeGetByKeyFinder) GetByKey(_ context.Context, _ tenant.TenantID, _ string) (*projectview.ProjectDetail, error) {
	return f.detail, f.err
}

func testCaseFileSummary() casefilelist.CaseFileSummary {
	cov := 85.5
	return casefilelist.CaseFileSummary{
		ID:               "22222222-2222-2222-2222-222222222222",
		ProjectID:        "11111111-1111-1111-1111-111111111111",
		CommitSHA:        "abc123",
		Branch:           "main",
		StartedAt:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		VerdictOutcome:   "pass",
		TotalFindings:    10,
		NewFindings:      2,
		ExistingFindings: 8,
		ResolvedFindings: 3,
		CoveragePercent:  &cov,
		CreatedAt:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func testCaseFileDetail() *casefileget.CaseFileDetail {
	cov := 85.5
	return &casefileget.CaseFileDetail{
		ID:               "22222222-2222-2222-2222-222222222222",
		ProjectID:        "11111111-1111-1111-1111-111111111111",
		CommitSHA:        "abc123",
		Branch:           "main",
		StartedAt:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		VerdictOutcome:   "pass",
		TotalFindings:    10,
		NewFindings:      2,
		ExistingFindings: 8,
		ResolvedFindings: 3,
		CoveragePercent:  &cov,
		CreatedAt:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Evidences: []casefileget.EvidenceSummary{
			{ID: "33333333-3333-3333-3333-333333333333", Subtype: "code_quality", Source: "golangci-lint"},
		},
		Rulings: []casefileget.RulingView{
			{Subtype: "code_quality", Passed: true, Detail: "0 findings"},
		},
	}
}

func testFindingView() findinglist.FindingView {
	return findinglist.FindingView{
		Tool:          "golangci-lint",
		RuleID:        "errcheck",
		Severity:      "warning",
		FilePath:      "main.go",
		Line:          10,
		Message:       "unchecked error",
		FingerprintID: "fp1",
		Status:        "new",
		Source:        "golangci-lint",
		CommitSHA:     "abc123",
		ProjectKey:    "core",
		CaseFileID:    "22222222-2222-2222-2222-222222222222",
	}
}

type fakeFileCoverageSaver struct {
	savedCaseFileID string
	savedEntries    []evidencedto.FileCoverage
	err             error
}

func (f *fakeFileCoverageSaver) Save(_ context.Context, caseFileID string, entries []evidencedto.FileCoverage) error {
	f.savedCaseFileID = caseFileID
	f.savedEntries = entries
	return f.err
}

type fakeBaselineFinder struct {
	detail *projectview.ProjectDetail
	err    error
}

func (f *fakeBaselineFinder) GetByKey(_ context.Context, _ tenant.TenantID, _ string) (*projectview.ProjectDetail, error) {
	return f.detail, f.err
}
