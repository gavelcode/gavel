package ingestcoverage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/casefile/ingestcoverage"
)

func TestNewCommand(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		format  string
		source  string
		wantErr bool
	}{
		{
			name:   "valid command",
			data:   []byte("TN:\n"),
			format: "lcov",
			source: "go-test",
		},
		{
			name:    "empty data rejected",
			data:    nil,
			format:  "lcov",
			source:  "go-test",
			wantErr: true,
		},
		{
			name:    "empty format rejected",
			data:    []byte("TN:\n"),
			format:  "",
			source:  "go-test",
			wantErr: true,
		},
		{
			name:    "blank format rejected",
			data:    []byte("TN:\n"),
			format:  "  ",
			source:  "go-test",
			wantErr: true,
		},
		{
			name:    "empty source rejected",
			data:    []byte("TN:\n"),
			format:  "lcov",
			source:  "",
			wantErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cmd, err := ingestcoverage.NewCommand(testCase.data, testCase.format, testCase.source)

			if testCase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ingestcoverage.ErrInvalidCommand)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.data, cmd.Data())
			assert.Equal(t, testCase.format, cmd.Format())
			assert.Equal(t, testCase.source, cmd.Source())
		})
	}
}

func TestNewCommandDefensiveCopyData(t *testing.T) {
	input := []byte("TN:\n")
	cmd, err := ingestcoverage.NewCommand(input, "lcov", "go-test")
	require.NoError(t, err)

	input[0] = 'X'

	assert.Equal(t, []byte("TN:\n"), cmd.Data(),
		"command copies input data; mutating input does not change command")

	got := cmd.Data()
	got[0] = 'Y'
	assert.Equal(t, []byte("TN:\n"), cmd.Data(),
		"command copies data on each get; mutating returned slice does not change command")
}
