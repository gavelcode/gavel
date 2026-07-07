package removeproject_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/gavelspace/removeproject"
)

func TestNewCommand(t *testing.T) {
	tests := []struct {
		name           string
		gavelspaceName string
		projectID      string
		wantErr        bool
	}{
		{name: "valid", gavelspaceName: "monorepo", projectID: "proj-1"},
		{name: "empty gavelspace name rejected", gavelspaceName: "", projectID: "proj-1", wantErr: true},
		{name: "blank gavelspace name rejected", gavelspaceName: "   ", projectID: "proj-1", wantErr: true},
		{name: "empty project id rejected", gavelspaceName: "monorepo", projectID: "", wantErr: true},
		{name: "blank project id rejected", gavelspaceName: "monorepo", projectID: "   ", wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cmd, err := removeproject.NewCommand(testTenant, testCase.gavelspaceName, testCase.projectID)

			if testCase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, removeproject.ErrInvalidCommand)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.gavelspaceName, cmd.GavelspaceID())
			assert.Equal(t, testCase.projectID, cmd.ProjectID())
		})
	}
}

func TestNewCommandRejectsEmptyTenant(t *testing.T) {
	_, err := removeproject.NewCommand("", "monorepo", "proj-1")
	require.Error(t, err)
	assert.ErrorIs(t, err, removeproject.ErrInvalidCommand)
}
