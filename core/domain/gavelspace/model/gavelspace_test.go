package model_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/gavelspace/model"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
)

var (
	testTime   = time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	testTenant = tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))
)

func TestNewGavelspace(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantName  string
		wantError bool
	}{
		{name: "valid name", input: "my-workspace", wantName: "my-workspace"},
		{name: "empty name rejected", input: "", wantError: true},
		{name: "whitespace-only rejected", input: "   ", wantError: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			gspace, err := model.NewGavelspace(testTenant, tcase.input)

			if tcase.wantError {
				require.Error(t, err)
				assert.ErrorIs(t, err, model.ErrInvalidGavelspace)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tcase.wantName, gspace.ID().String())
			assert.True(t, testTenant.Equal(gspace.TenantID()))
			assert.Empty(t, gspace.Projects())
			assert.Empty(t, gspace.Events())
		})
	}
}

func TestGavelspaceAddProject(t *testing.T) {
	gspace, err := model.NewGavelspace(testTenant, "workspace")
	require.NoError(t, err)

	projID := projectmodel.NewProjectID(uuid.New())
	ref, err := model.NewProjectRef(projID, "//src/...")
	require.NoError(t, err)

	err = gspace.AddProject(ref, testTime)
	require.NoError(t, err)

	projects := gspace.Projects()
	require.Len(t, projects, 1)
	assert.True(t, projID.Equal(projects[0].ID()))
	assert.Equal(t, "//src/...", projects[0].TargetPattern())

	events := gspace.Events()
	require.Len(t, events, 1)
	added, ok := events[0].(model.ProjectAdded)
	require.True(t, ok)
	assert.Equal(t, "workspace", added.GavelspaceID().String())
	assert.True(t, projID.Equal(added.ProjectID()))
	assert.Equal(t, "//src/...", added.TargetPattern())
}

func TestGavelspaceAddProjectDuplicateTargetPattern(t *testing.T) {
	gspace, err := model.NewGavelspace(testTenant, "workspace")
	require.NoError(t, err)

	ref1, err := model.NewProjectRef(projectmodel.NewProjectID(uuid.New()), "//src/...")
	require.NoError(t, err)
	ref2, err := model.NewProjectRef(projectmodel.NewProjectID(uuid.New()), "//src/...")
	require.NoError(t, err)

	require.NoError(t, gspace.AddProject(ref1, testTime))

	err = gspace.AddProject(ref2, testTime)
	require.Error(t, err)
	assert.ErrorIs(t, err, model.ErrDuplicateTargetPattern)
}

func TestGavelspaceAddProjectDifferentPatterns(t *testing.T) {
	gspace, err := model.NewGavelspace(testTenant, "workspace")
	require.NoError(t, err)

	ref1, err := model.NewProjectRef(projectmodel.NewProjectID(uuid.New()), "//src/...")
	require.NoError(t, err)
	ref2, err := model.NewProjectRef(projectmodel.NewProjectID(uuid.New()), "//lib/...")
	require.NoError(t, err)

	require.NoError(t, gspace.AddProject(ref1, testTime))
	require.NoError(t, gspace.AddProject(ref2, testTime))

	projects := gspace.Projects()
	require.Len(t, projects, 2)
}

func TestGavelspaceRemoveProject(t *testing.T) {
	gspace, err := model.NewGavelspace(testTenant, "workspace")
	require.NoError(t, err)

	projID := projectmodel.NewProjectID(uuid.New())
	ref, err := model.NewProjectRef(projID, "//src/...")
	require.NoError(t, err)
	require.NoError(t, gspace.AddProject(ref, testTime))
	gspace.ClearEvents()

	err = gspace.RemoveProject(projID, testTime)
	require.NoError(t, err)

	assert.Empty(t, gspace.Projects())

	events := gspace.Events()
	require.Len(t, events, 1)
	removed, ok := events[0].(model.ProjectRemoved)
	require.True(t, ok)
	assert.Equal(t, "workspace", removed.GavelspaceID().String())
	assert.True(t, projID.Equal(removed.ProjectID()))
}

func TestGavelspaceRemoveProjectNotFound(t *testing.T) {
	gspace, err := model.NewGavelspace(testTenant, "workspace")
	require.NoError(t, err)

	err = gspace.RemoveProject(projectmodel.NewProjectID(uuid.New()), testTime)
	require.Error(t, err)
	assert.ErrorIs(t, err, model.ErrProjectNotFound)
}

func TestGavelspaceProjectsDefensiveCopy(t *testing.T) {
	gspace, err := model.NewGavelspace(testTenant, "workspace")
	require.NoError(t, err)

	projID := projectmodel.NewProjectID(uuid.New())
	ref, err := model.NewProjectRef(projID, "//src/...")
	require.NoError(t, err)
	require.NoError(t, gspace.AddProject(ref, testTime))

	projects := gspace.Projects()
	projects[0] = model.ProjectRef{}

	assert.True(t, projID.Equal(gspace.Projects()[0].ID()), "mutating returned slice must not affect aggregate")
}

func TestReconstituteGavelspace(t *testing.T) {
	name, err := model.NewGavelspaceID("workspace")
	require.NoError(t, err)

	t.Run("valid with pre-existing projects", func(t *testing.T) {
		ref1, err := model.NewProjectRef(projectmodel.NewProjectID(uuid.New()), "//src/...")
		require.NoError(t, err)
		ref2, err := model.NewProjectRef(projectmodel.NewProjectID(uuid.New()), "//lib/...")
		require.NoError(t, err)

		gspace, err := model.ReconstituteGavelspace(name, testTenant, []model.ProjectRef{ref1, ref2})
		require.NoError(t, err)

		assert.Equal(t, "workspace", gspace.ID().String())
		projects := gspace.Projects()
		require.Len(t, projects, 2)
		assert.Empty(t, gspace.Events())
	})

	t.Run("defensive copy of projects", func(t *testing.T) {
		projID := projectmodel.NewProjectID(uuid.New())
		ref, err := model.NewProjectRef(projID, "//src/...")
		require.NoError(t, err)

		input := []model.ProjectRef{ref}
		gspace, err := model.ReconstituteGavelspace(name, testTenant, input)
		require.NoError(t, err)

		input[0] = model.ProjectRef{}

		assert.True(t, projID.Equal(gspace.Projects()[0].ID()))
	})

	t.Run("duplicate target pattern rejected", func(t *testing.T) {
		ref1, err := model.NewProjectRef(projectmodel.NewProjectID(uuid.New()), "//src/...")
		require.NoError(t, err)
		ref2, err := model.NewProjectRef(projectmodel.NewProjectID(uuid.New()), "//src/...")
		require.NoError(t, err)

		_, err = model.ReconstituteGavelspace(name, testTenant, []model.ProjectRef{ref1, ref2})
		require.Error(t, err)
		assert.ErrorIs(t, err, model.ErrDuplicateTargetPattern)
	})
}

func TestGavelspaceClearEvents(t *testing.T) {
	gspace, err := model.NewGavelspace(testTenant, "workspace")
	require.NoError(t, err)

	ref, err := model.NewProjectRef(projectmodel.NewProjectID(uuid.New()), "//src/...")
	require.NoError(t, err)
	require.NoError(t, gspace.AddProject(ref, testTime))
	require.NotEmpty(t, gspace.Events())

	gspace.ClearEvents()

	assert.Empty(t, gspace.Events())
}

func TestNewGavelspaceRejectsZeroTenant(t *testing.T) {
	_, err := model.NewGavelspace(tenant.TenantID{}, "workspace")
	require.Error(t, err)
	assert.ErrorIs(t, err, model.ErrInvalidGavelspace)
}
