package updatelanguages_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/project/updatelanguages"
)

func TestNewCommand(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		languages []string
		wantErr   bool
	}{
		{name: "valid with languages", projectID: "proj-1", languages: []string{"java", "go"}},
		{name: "valid empty list", projectID: "proj-1", languages: nil},
		{name: "empty project id rejected", projectID: "", languages: []string{"java"}, wantErr: true},
		{name: "blank project id rejected", projectID: "   ", languages: []string{"java"}, wantErr: true},
		{name: "invalid language rejected", projectID: "proj-1", languages: []string{""}, wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cmd, err := updatelanguages.NewCommand(testTenant.String(), testCase.projectID, testCase.languages)

			if testCase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, updatelanguages.ErrInvalidCommand)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.projectID, cmd.ProjectID())
			assert.Equal(t, len(testCase.languages), len(cmd.Languages()))
		})
	}
}

func TestNewCommandDefensiveCopy(t *testing.T) {
	input := []string{"java"}
	cmd, err := updatelanguages.NewCommand(testTenant.String(), "proj-1", input)
	require.NoError(t, err)

	input[0] = "go"

	assert.Equal(t, "java", cmd.Languages()[0].String(),
		"command parses languages defensively; mutating input does not change command")
}
