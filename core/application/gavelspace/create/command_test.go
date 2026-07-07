package create_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/gavelspace/create"
)

func TestNewCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid", input: "monorepo"},
		{name: "empty name rejected", input: "", wantErr: true},
		{name: "blank name rejected", input: "   ", wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cmd, err := create.NewCommand(testTenant, testCase.input)

			if testCase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, create.ErrInvalidCommand)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.input, cmd.Name())
		})
	}
}

func TestNewCommandRejectsEmptyTenant(t *testing.T) {
	_, err := create.NewCommand("", "monorepo")
	require.Error(t, err)
	assert.ErrorIs(t, err, create.ErrInvalidCommand)
}
