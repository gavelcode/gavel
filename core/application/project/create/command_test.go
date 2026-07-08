package create_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/project/create"
)

func TestNewCommand(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		projectName   string
		targetPattern string
		wantErr       bool
	}{
		{name: "valid with target pattern", key: "svc", projectName: "svc", targetPattern: "//svc/..."},
		{name: "empty target pattern rejected", key: "svc", projectName: "svc", targetPattern: "", wantErr: true},
		{name: "empty key rejected", key: "", projectName: "svc", targetPattern: "//svc/...", wantErr: true},
		{name: "blank key rejected", key: "   ", projectName: "svc", targetPattern: "//svc/...", wantErr: true},
		{name: "empty name rejected", key: "svc", projectName: "", targetPattern: "//svc/...", wantErr: true},
		{name: "blank name rejected", key: "svc", projectName: "   ", targetPattern: "//svc/...", wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cmd, err := create.NewCommand(testTenant.String(), testCase.key, testCase.projectName, testCase.targetPattern)

			if testCase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, create.ErrInvalidCommand)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.key, cmd.Key())
			assert.Equal(t, testCase.projectName, cmd.Name())
			assert.Equal(t, testCase.targetPattern, cmd.TargetPattern())
		})
	}
}
