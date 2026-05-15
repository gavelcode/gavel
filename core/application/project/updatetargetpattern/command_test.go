package updatetargetpattern_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/project/updatetargetpattern"
)

func TestNewCommand(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		pattern   string
		wantErr   bool
	}{
		{name: "valid", projectID: "proj-1", pattern: "//svc/..."},
		{name: "empty project id rejected", projectID: "", pattern: "//svc/...", wantErr: true},
		{name: "blank project id rejected", projectID: "   ", pattern: "//svc/...", wantErr: true},
		{name: "empty pattern rejected", projectID: "proj-1", pattern: "", wantErr: true},
		{name: "blank pattern rejected", projectID: "proj-1", pattern: "   ", wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cmd, err := updatetargetpattern.NewCommand(testCase.projectID, testCase.pattern)

			if testCase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, updatetargetpattern.ErrInvalidCommand)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.projectID, cmd.ProjectID())
			assert.Equal(t, testCase.pattern, cmd.TargetPattern())
		})
	}
}
