package source_test

import (
	"context"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	findinglist "github.com/usegavel/gavel/core/application/casefile/listfindings"
	projectgetbykey "github.com/usegavel/gavel/core/application/project/getbykey"
	"github.com/usegavel/gavel/core/application/project/projectview"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	"github.com/usegavel/gavel/core/domain/shared/failure"
	auth "github.com/usegavel/gavel/core/userinterface/api/v1/server/httpx/auth"
)

type fakeBlobs struct {
	content []byte
	ct      string
	err     error
}

func (f *fakeBlobs) Fetch(_ context.Context, _, _, _ string) ([]byte, string, error) {
	return f.content, f.ct, f.err
}

func (f *fakeBlobs) Save(_ context.Context, _, _, _ string, _ []byte, _ string) error {
	return nil
}

type savedFile struct {
	projectID   string
	commitSHA   string
	path        string
	contentType string
	content     []byte
}

type trackingBlobs struct {
	fakeBlobs
	saved []savedFile
}

func (t *trackingBlobs) Save(_ context.Context, projectID, commitSHA, path string, content []byte, contentType string) error {
	t.saved = append(t.saved, savedFile{
		projectID:   projectID,
		commitSHA:   commitSHA,
		path:        path,
		contentType: contentType,
		content:     content,
	})
	return nil
}

type notFoundBlobs struct{}

func (n *notFoundBlobs) Fetch(_ context.Context, _, _, _ string) ([]byte, string, error) {
	return nil, "", failure.New("not found", failure.NotFound)
}

func (n *notFoundBlobs) Save(_ context.Context, _, _, _ string, _ []byte, _ string) error {
	return nil
}

type fakeCoverageFetcher struct {
	result *evidencedto.FileCoverage
	err    error
}

func (f *fakeCoverageFetcher) Fetch(_ context.Context, _, _ string) (*evidencedto.FileCoverage, error) {
	return f.result, f.err
}

type fakeFindingFetcher struct {
	items []findinglist.FindingView
	err   error
}

func (f *fakeFindingFetcher) ListByFile(_ context.Context, _, _, _ string) ([]findinglist.FindingView, error) {
	return f.items, f.err
}

func fakeProjectResolver(projectID string) *projectgetbykey.Handler {
	return projectgetbykey.NewHandler(&fakeProjectFinder{id: projectID})
}

type fakeProjectFinder struct {
	id string
}

func (f *fakeProjectFinder) GetByKey(_ context.Context, _ tenant.TenantID, _ string) (*projectview.ProjectDetail, error) {
	return &projectview.ProjectDetail{ID: f.id, Key: "core", Name: "core"}, nil
}

type notFoundProjectFinder struct{}

func (n *notFoundProjectFinder) GetByKey(_ context.Context, _ tenant.TenantID, _ string) (*projectview.ProjectDetail, error) {
	return nil, failure.New("not found", failure.NotFound)
}

func notFoundProjectResolver() *projectgetbykey.Handler {
	return projectgetbykey.NewHandler(&notFoundProjectFinder{})
}

const testTenant = "22222222-2222-2222-2222-222222222222"

func authContext() context.Context {
	return auth.WithPrincipal(context.Background(), &auth.Principal{TenantID: testTenant})
}
