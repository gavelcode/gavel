package loadgavelspace_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/gavelspace/loadgavelspace"
	gavelspacemodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/model/archpolicy"
)

type fakeFinder struct {
	gavelspace gavelspacemodel.Gavelspace
	projects   []projectmodel.Project
	err        error
}

func (f *fakeFinder) LoadFromConfig(_ string) (gavelspacemodel.Gavelspace, []projectmodel.Project, error) {
	return f.gavelspace, f.projects, f.err
}

func TestExecute_ReturnsGavelspaceAndProjects(t *testing.T) {
	gavelspace, err := gavelspacemodel.NewGavelspace("myrepo")
	require.NoError(t, err)
	gavelspace.SetServerConfig(gavelspacemodel.NewServerConfig("https://gavel.dev", "tok"))

	p, err := projectmodel.NewProject("core", "core", "//core/...")
	require.NoError(t, err)

	finder := &fakeFinder{gavelspace: gavelspace, projects: []projectmodel.Project{p}}
	h := loadgavelspace.NewHandler(finder)

	q, err := loadgavelspace.NewQuery("/path/to/gavel.yaml")
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), q)

	require.NoError(t, err)
	assert.Equal(t, "myrepo", result.Gavelspace.ID().String())
	assert.True(t, result.Gavelspace.ServerConfig().IsConfigured())
	assert.Len(t, result.Projects, 1)
	assert.Equal(t, "core", result.Projects[0].Name())
}

func TestExecute_FinderError(t *testing.T) {
	finder := &fakeFinder{err: fmt.Errorf("file not found")}
	h := loadgavelspace.NewHandler(finder)

	q, err := loadgavelspace.NewQuery("/bad/path")
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), q)

	assert.Error(t, err)
}

func TestNewQuery_EmptyPath(t *testing.T) {
	_, err := loadgavelspace.NewQuery("")
	assert.ErrorIs(t, err, loadgavelspace.ErrInvalidQuery)
}

type fakeArchLoader struct {
	policy archpolicy.Policy
	err    error
}

func (f *fakeArchLoader) LoadPolicy(_ string) (archpolicy.Policy, error) {
	return f.policy, f.err
}

type fakeProjectSaver struct {
	saved []projectmodel.Project
}

func (f *fakeProjectSaver) Save(_ context.Context, p projectmodel.Project) error {
	f.saved = append(f.saved, p)
	return nil
}

func (f *fakeProjectSaver) FindByKey(_ context.Context, key string) (projectmodel.Project, error) {
	for _, p := range f.saved {
		if p.Key() == key {
			return p, nil
		}
	}
	return projectmodel.Project{}, fmt.Errorf("not found: %s", key)
}

type failingProjectSaver struct {
	err error
}

func (f *failingProjectSaver) Save(_ context.Context, _ projectmodel.Project) error {
	return f.err
}

func (f *failingProjectSaver) FindByKey(_ context.Context, _ string) (projectmodel.Project, error) {
	return projectmodel.Project{}, fmt.Errorf("not found")
}

func TestExecute_AppliesArchPolicyWhenLoaderProvided(t *testing.T) {
	project, err := projectmodel.NewProject("core", "core", "//core/...")
	require.NoError(t, err)

	layer, err := archpolicy.NewLayer("domain", []string{"core/domain/"})
	require.NoError(t, err)
	policy, err := archpolicy.NewPolicy([]archpolicy.Layer{layer}, nil, false)
	require.NoError(t, err)

	gavelspace, err := gavelspacemodel.NewGavelspace("test")
	require.NoError(t, err)
	finder := &fakeFinder{gavelspace: gavelspace, projects: []projectmodel.Project{project}}
	loader := &fakeArchLoader{policy: policy}

	h := loadgavelspace.NewHandler(finder, loadgavelspace.WithArchPolicyLoader(loader))

	q, err := loadgavelspace.NewQuery("/config.yaml", loadgavelspace.WithWorkspace("/workspace"))
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), q)
	require.NoError(t, err)

	assert.NotNil(t, result.Projects[0].Policy())
}

func TestExecute_SavesProjectsWhenSaverProvided(t *testing.T) {
	project, err := projectmodel.NewProject("core", "core", "//core/...")
	require.NoError(t, err)

	gavelspace, err := gavelspacemodel.NewGavelspace("test")
	require.NoError(t, err)
	finder := &fakeFinder{gavelspace: gavelspace, projects: []projectmodel.Project{project}}
	saver := &fakeProjectSaver{}

	h := loadgavelspace.NewHandler(finder, loadgavelspace.WithProjectSaver(saver))

	q, err := loadgavelspace.NewQuery("/config.yaml")
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), q)
	require.NoError(t, err)

	require.Len(t, saver.saved, 1)
	assert.Equal(t, "core", saver.saved[0].Name())
}

func TestExecute_FiltersProjectByName(t *testing.T) {
	coreProject, err := projectmodel.NewProject("core", "core", "//core/...")
	require.NoError(t, err)
	webProject, err := projectmodel.NewProject("web", "web", "//web/...")
	require.NoError(t, err)

	gavelspace, err := gavelspacemodel.NewGavelspace("test")
	require.NoError(t, err)
	finder := &fakeFinder{gavelspace: gavelspace, projects: []projectmodel.Project{coreProject, webProject}}

	h := loadgavelspace.NewHandler(finder)

	q, err := loadgavelspace.NewQuery("/config.yaml", loadgavelspace.WithProjectFilter("web"))
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), q)
	require.NoError(t, err)

	require.Len(t, result.Projects, 1)
	assert.Equal(t, "web", result.Projects[0].Name())
}

func TestExecute_FilterNotFoundReturnsError(t *testing.T) {
	project, err := projectmodel.NewProject("core", "core", "//core/...")
	require.NoError(t, err)

	gavelspace, err := gavelspacemodel.NewGavelspace("test")
	require.NoError(t, err)
	finder := &fakeFinder{gavelspace: gavelspace, projects: []projectmodel.Project{project}}

	h := loadgavelspace.NewHandler(finder)

	q, err := loadgavelspace.NewQuery("/config.yaml", loadgavelspace.WithProjectFilter("nonexistent"))
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), q)
	assert.ErrorIs(t, err, loadgavelspace.ErrInvalidQuery)
}

func TestNewHandlerPanicsOnNilFinder(t *testing.T) {
	assert.Panics(t, func() { loadgavelspace.NewHandler(nil) })
}

func TestExecute_WithLoggerOption(t *testing.T) {
	gavelspace, err := gavelspacemodel.NewGavelspace("test")
	require.NoError(t, err)
	p, err := projectmodel.NewProject("core", "core", "//core/...")
	require.NoError(t, err)

	finder := &fakeFinder{gavelspace: gavelspace, projects: []projectmodel.Project{p}}
	h := loadgavelspace.NewHandler(finder, loadgavelspace.WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))))

	q, err := loadgavelspace.NewQuery("/config.yaml")
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), q)
	require.NoError(t, err)
	assert.Len(t, result.Projects, 1)
}

func TestExecute_ProjectSaverErrorPropagated(t *testing.T) {
	gavelspace, err := gavelspacemodel.NewGavelspace("test")
	require.NoError(t, err)
	p, err := projectmodel.NewProject("core", "core", "//core/...")
	require.NoError(t, err)

	finder := &fakeFinder{gavelspace: gavelspace, projects: []projectmodel.Project{p}}
	saver := &failingProjectSaver{err: fmt.Errorf("disk full")}

	h := loadgavelspace.NewHandler(finder, loadgavelspace.WithProjectSaver(saver))

	q, err := loadgavelspace.NewQuery("/config.yaml")
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), q)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save project")
}

func TestExecute_ArchLoaderErrorSkipped(t *testing.T) {
	p, err := projectmodel.NewProject("core", "core", "//core/...")
	require.NoError(t, err)

	gavelspace, err := gavelspacemodel.NewGavelspace("test")
	require.NoError(t, err)
	finder := &fakeFinder{gavelspace: gavelspace, projects: []projectmodel.Project{p}}
	loader := &fakeArchLoader{err: fmt.Errorf("file not found")}

	h := loadgavelspace.NewHandler(finder, loadgavelspace.WithArchPolicyLoader(loader))

	q, err := loadgavelspace.NewQuery("/config.yaml", loadgavelspace.WithWorkspace("/workspace"))
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), q)
	require.NoError(t, err)
	assert.Nil(t, result.Projects[0].Policy())
}
