package getbaseline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/project/getbaseline"
	"github.com/usegavel/gavel/core/application/project/projectview"
	"github.com/usegavel/gavel/core/domain/shared/failure"
)

func TestNewQuery(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		branch  string
		wantErr bool
	}{
		{name: "validQuery", key: "core", branch: "main"},
		{name: "validEmptyBranch", key: "core", branch: ""},
		{name: "emptyKey", key: "", branch: "main", wantErr: true},
		{name: "whitespaceKey", key: "  ", branch: "main", wantErr: true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			query, err := getbaseline.NewQuery(testTenant.String(), testCase.key, testCase.branch)
			if testCase.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, getbaseline.ErrInvalidQuery)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.key, query.Key())
			assert.Equal(t, testCase.branch, query.Branch())
		})
	}
}

func TestNewHandlerPanicsOnNilFinder(t *testing.T) {
	repo := &fakeProjectRepo{}
	assert.Panics(t, func() { getbaseline.NewHandler(nil, repo) })
}

func TestNewHandlerPanicsOnNilRepo(t *testing.T) {
	finder := &fakeProjectFinder{}
	assert.Panics(t, func() { getbaseline.NewHandler(finder, nil) })
}

func TestExecuteReturnsBaseline(t *testing.T) {
	proj := seedProjectWithBaseline(t, "main", []string{"fp1", "fp2"}, []string{"a1"})
	finder := &fakeProjectFinder{detail: projectDetail("core", proj)}
	repo := &fakeProjectRepo{project: proj}
	h := getbaseline.NewHandler(finder, repo)

	query, err := getbaseline.NewQuery(testTenant.String(), "core", "main")
	require.NoError(t, err)

	res, err := h.Execute(context.Background(), query)
	require.NoError(t, err)
	assert.Equal(t, []string{"fp1", "fp2"}, res.Fingerprints)
	assert.Equal(t, []string{"a1"}, res.ArchIDs)
	assert.True(t, res.HasPrevious)
}

func TestExecuteProjectNotFoundByKey(t *testing.T) {
	finder := &fakeProjectFinder{err: failure.New("not found", failure.NotFound)}
	repo := &fakeProjectRepo{}
	h := getbaseline.NewHandler(finder, repo)

	query, err := getbaseline.NewQuery(testTenant.String(), "missing", "main")
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), query)
	assert.Equal(t, failure.NotFound, failure.Of(err))
}

func TestExecuteEmptyBaselineReturnsEmptyResult(t *testing.T) {
	proj := seedProjectWithoutBaseline(t)
	finder := &fakeProjectFinder{detail: projectDetail("core", proj)}
	repo := &fakeProjectRepo{project: proj}
	h := getbaseline.NewHandler(finder, repo)

	query, err := getbaseline.NewQuery(testTenant.String(), "core", "main")
	require.NoError(t, err)

	res, err := h.Execute(context.Background(), query)
	require.NoError(t, err)
	assert.Equal(t, []string{}, res.Fingerprints)
	assert.Equal(t, []string{}, res.ArchIDs)
	assert.False(t, res.HasPrevious)
}

func TestExecuteInvalidProjectIDFromFinder(t *testing.T) {
	finder := &fakeProjectFinder{detail: &projectview.ProjectDetail{
		ID:            "not-a-uuid",
		Key:           "core",
		DefaultBranch: "main",
	}}
	repo := &fakeProjectRepo{}
	h := getbaseline.NewHandler(finder, repo)

	query, err := getbaseline.NewQuery(testTenant.String(), "core", "main")
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), query)
	assert.Error(t, err)
}

func TestExecuteProjectNotFoundByID(t *testing.T) {
	proj := seedProjectWithoutBaseline(t)
	finder := &fakeProjectFinder{detail: projectDetail("core", proj)}
	repo := &fakeProjectRepo{err: failure.New("not found", failure.NotFound)}
	h := getbaseline.NewHandler(finder, repo)

	query, err := getbaseline.NewQuery(testTenant.String(), "core", "main")
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), query)
	assert.Error(t, err)
}

func TestExecuteUsesDefaultBranchWhenQueryBranchEmpty(t *testing.T) {
	proj := seedProjectWithBaseline(t, "develop", []string{"fp-dev"}, nil)
	finder := &fakeProjectFinder{detail: projectDetailWithBranch("core", "develop", proj)}
	repo := &fakeProjectRepo{project: proj}
	h := getbaseline.NewHandler(finder, repo)

	query, err := getbaseline.NewQuery(testTenant.String(), "core", "")
	require.NoError(t, err)

	res, err := h.Execute(context.Background(), query)
	require.NoError(t, err)
	assert.Equal(t, []string{"fp-dev"}, res.Fingerprints)
}
