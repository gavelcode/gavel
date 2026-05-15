package preparebaseline_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/project/preparebaseline"
)

func TestExecuteFetchesBaselineFromServer(t *testing.T) {
	projRepo := newFakeProjectRepo()
	cfRepo := newFakeCaseFileRepo()
	project := seedProject(t, projRepo, "core")

	fetcher := &fakeFetcher{
		baselines: map[string]*preparebaseline.RemoteBaseline{
			"core|main": {Fingerprints: []string{"fp-1", "fp-2"}, HasPrevious: true},
		},
	}

	handler := preparebaseline.NewHandler(projRepo, cfRepo, preparebaseline.WithFetcher(fetcher))
	cmd := mustCommand(t, []preparebaseline.ProjectInput{{Name: "core", DefaultBranch: "main"}})

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, result.Baselines, 1)
	assert.Equal(t, "core", result.Baselines[0].ProjectName)
	assert.Equal(t, 2, result.Baselines[0].FingerprintCount)
	assert.True(t, result.Baselines[0].HasPrevious)
	assert.Equal(t, "server", result.Baselines[0].Source)

	saved := projRepo.lastSaved()
	bl := saved.Baseline("main")
	assert.Equal(t, []string{"fp-1", "fp-2"}, bl.Fingerprints())

	assert.Equal(t, 2, cfRepo.preloadedCount(project.ID(), "main"))
}

func TestExecuteUsesLocalBaselineWhenNoFetcher(t *testing.T) {
	projRepo := newFakeProjectRepo()
	cfRepo := newFakeCaseFileRepo()
	project := seedProject(t, projRepo, "local")
	project.UpdateBaseline("main", []string{"fp-local"}, []string{"arch-1"}, nil, nil)
	require.NoError(t, projRepo.Save(context.Background(), project))

	handler := preparebaseline.NewHandler(projRepo, cfRepo)
	cmd := mustCommand(t, []preparebaseline.ProjectInput{{Name: "local", DefaultBranch: "main"}})

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, result.Baselines, 1)
	assert.Equal(t, 1, result.Baselines[0].FingerprintCount)
	assert.True(t, result.Baselines[0].HasPrevious)
	assert.Equal(t, "local", result.Baselines[0].Source)

	assert.Equal(t, 1, cfRepo.preloadedCount(project.ID(), "main"))
}

func TestExecuteNoBaselineFirstRun(t *testing.T) {
	projRepo := newFakeProjectRepo()
	cfRepo := newFakeCaseFileRepo()
	seedProject(t, projRepo, "new")

	handler := preparebaseline.NewHandler(projRepo, cfRepo)
	cmd := mustCommand(t, []preparebaseline.ProjectInput{{Name: "new", DefaultBranch: "main"}})

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, result.Baselines, 1)
	assert.False(t, result.Baselines[0].HasPrevious)
	assert.Equal(t, 0, cfRepo.preloadedCount(projectIDByName(projRepo, "new"), "main"))
}

func TestExecuteRemoteBaselineSaveErrorLogsWarning(t *testing.T) {
	projRepo := newFakeProjectRepo()
	cfRepo := newFakeCaseFileRepo()
	seedProject(t, projRepo, "core")

	projRepo.saveErr = errors.New("disk full")

	fetcher := &fakeFetcher{
		baselines: map[string]*preparebaseline.RemoteBaseline{
			"core|main": {Fingerprints: []string{"fp-1"}, HasPrevious: true},
		},
	}

	handler := preparebaseline.NewHandler(projRepo, cfRepo, preparebaseline.WithFetcher(fetcher))
	cmd := mustCommand(t, []preparebaseline.ProjectInput{{Name: "core", DefaultBranch: "main"}})

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, result.Baselines, 1)
	assert.Equal(t, "server", result.Baselines[0].Source)
}

func TestExecuteRemoteBaselineSkipsInvalidFingerprints(t *testing.T) {
	projRepo := newFakeProjectRepo()
	cfRepo := newFakeCaseFileRepo()
	project := seedProject(t, projRepo, "core")

	fetcher := &fakeFetcher{
		baselines: map[string]*preparebaseline.RemoteBaseline{
			"core|main": {Fingerprints: []string{"valid-fp", "", "  "}, HasPrevious: true},
		},
	}

	handler := preparebaseline.NewHandler(projRepo, cfRepo, preparebaseline.WithFetcher(fetcher))
	cmd := mustCommand(t, []preparebaseline.ProjectInput{{Name: "core", DefaultBranch: "main"}})

	result, err := handler.Execute(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, result.Baselines, 1)
	assert.Equal(t, 3, result.Baselines[0].FingerprintCount)
	assert.Equal(t, 1, cfRepo.preloadedCount(project.ID(), "main"))
}

func TestNewHandlerPanicsOnNilRepos(t *testing.T) {
	projRepo := newFakeProjectRepo()
	cfRepo := newFakeCaseFileRepo()
	assert.Panics(t, func() { preparebaseline.NewHandler(nil, cfRepo) })
	assert.Panics(t, func() { preparebaseline.NewHandler(projRepo, nil) })
}

func TestNewCommandRejectsEmptyProjectName(t *testing.T) {
	_, err := preparebaseline.NewCommand([]preparebaseline.ProjectInput{{Name: "", DefaultBranch: "main"}})
	assert.ErrorIs(t, err, preparebaseline.ErrInvalidCommand)
}

func mustCommand(t *testing.T, projects []preparebaseline.ProjectInput) preparebaseline.Command {
	t.Helper()
	cmd, err := preparebaseline.NewCommand(projects)
	require.NoError(t, err)
	return cmd
}
