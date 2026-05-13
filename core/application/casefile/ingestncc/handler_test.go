package ingestncc_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/ingestncc"
)

type fakePerLineParser struct {
	result map[string]map[int]int
	err    error
}

func (f fakePerLineParser) ParsePerLine(_ []byte) (map[string]map[int]int, error) {
	return f.result, f.err
}

func TestExecute_ComputesPatchCoverage(t *testing.T) {
	parser := fakePerLineParser{
		result: map[string]map[int]int{
			"main.go": {1: 1, 2: 1, 3: 0, 4: 1},
		},
	}
	changedLines := map[string][]int{
		"main.go": {1, 2, 3, 4},
	}

	h := ingestncc.NewHandler(parser)
	cmd, err := ingestncc.NewCommand([]byte("SF:main.go\n"), changedLines)
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, "new_code_coverage", result.Evidence.Subtype)
	assert.Equal(t, "gavel", result.Evidence.Source)
	assert.NotNil(t, result.Evidence.NewCodeCoverage)
	assert.Equal(t, 3, result.Evidence.NewCodeCoverage.CoveredLines)
	assert.Equal(t, 4, result.Evidence.NewCodeCoverage.CoverableLines)
	assert.InDelta(t, 75.0, result.Percent, 0.1)
}

func TestExecute_ParserError(t *testing.T) {
	parser := fakePerLineParser{err: fmt.Errorf("bad lcov")}

	h := ingestncc.NewHandler(parser)
	cmd, err := ingestncc.NewCommand([]byte("data"), map[string][]int{"main.go": {1}})
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), cmd)

	assert.ErrorIs(t, err, ingestncc.ErrParseFailed)
}

func TestNewCommand_EmptyLCOV(t *testing.T) {
	_, err := ingestncc.NewCommand(nil, map[string][]int{"f": {1}})
	assert.ErrorIs(t, err, ingestncc.ErrInvalidCommand)
}

func TestNewCommand_EmptyChangedLines(t *testing.T) {
	_, err := ingestncc.NewCommand([]byte("data"), nil)
	assert.ErrorIs(t, err, ingestncc.ErrInvalidCommand)
}

func TestExecute_NoOverlap_ReturnsZeroCoverage(t *testing.T) {
	parser := fakePerLineParser{
		result: map[string]map[int]int{
			"other.go": {1: 1, 2: 1, 3: 0},
		},
	}
	changedLines := map[string][]int{
		"main.go": {1, 2},
	}

	h := ingestncc.NewHandler(parser)
	cmd, err := ingestncc.NewCommand([]byte("data"), changedLines)
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, result.Evidence.NewCodeCoverage)
	assert.Equal(t, 0, result.Evidence.NewCodeCoverage.CoverableLines)
	assert.Equal(t, 0.0, result.Percent)
}
