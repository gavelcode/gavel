package ingestfindings_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/application/casefile/ingestfindings"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
)

func TestNewCommand(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		format  string
		source  string
		subtype string
		wantErr bool
	}{
		{
			name:    "valid command",
			data:    []byte("{}"),
			format:  "sarif",
			source:  "spotbugs",
			subtype: evidence.SubtypeCodeQuality.String(),
		},
		{
			name:    "empty data rejected",
			data:    nil,
			format:  "sarif",
			source:  "spotbugs",
			subtype: evidence.SubtypeCodeQuality.String(),
			wantErr: true,
		},
		{
			name:    "empty format rejected",
			data:    []byte("{}"),
			format:  "",
			source:  "spotbugs",
			subtype: evidence.SubtypeCodeQuality.String(),
			wantErr: true,
		},
		{
			name:    "blank format rejected",
			data:    []byte("{}"),
			format:  "  ",
			source:  "spotbugs",
			subtype: evidence.SubtypeCodeQuality.String(),
			wantErr: true,
		},
		{
			name:    "empty source rejected",
			data:    []byte("{}"),
			format:  "sarif",
			source:  "",
			subtype: evidence.SubtypeCodeQuality.String(),
			wantErr: true,
		},
		{
			name:    "unknown subtype rejected",
			data:    []byte("{}"),
			format:  "sarif",
			source:  "spotbugs",
			subtype: "nonsense",
			wantErr: true,
		},
		{
			name:    "coverage subtype rejected (not finding-based)",
			data:    []byte("{}"),
			format:  "sarif",
			source:  "spotbugs",
			subtype: evidence.SubtypeCoverage.String(),
			wantErr: true,
		},
		{
			name:    "license subtype rejected (not finding-based)",
			data:    []byte("{}"),
			format:  "sarif",
			source:  "spotbugs",
			subtype: evidence.SubtypeLicense.String(),
			wantErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cmd, err := ingestfindings.NewCommand(testCase.data, testCase.format, testCase.source, testCase.subtype)

			if testCase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ingestfindings.ErrInvalidCommand)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.data, cmd.Data())
			assert.Equal(t, testCase.format, cmd.Format())
			assert.Equal(t, testCase.source, cmd.Source())
			assert.Equal(t, testCase.subtype, cmd.Subtype().String())
		})
	}
}

func TestNewCommandDefensiveCopyData(t *testing.T) {
	input := []byte("original")
	cmd, err := ingestfindings.NewCommand(input, "sarif", "spotbugs", evidence.SubtypeCodeQuality.String())
	require.NoError(t, err)

	input[0] = 'X'

	assert.Equal(t, []byte("original"), cmd.Data(),
		"command copies input data; mutating input does not change command")

	got := cmd.Data()
	got[0] = 'Y'
	assert.Equal(t, []byte("original"), cmd.Data(),
		"command copies data on each get; mutating returned slice does not change command")
}
