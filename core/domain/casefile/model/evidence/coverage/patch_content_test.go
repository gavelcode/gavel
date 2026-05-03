package coverage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
)

func TestNewPatchContentFromChanges(t *testing.T) {
	tests := []struct {
		name          string
		changedLines  map[string][]int
		perLine       map[string]map[int]int
		wantCovered   int
		wantCoverable int
	}{
		{
			name:          "all changed lines covered",
			changedLines:  map[string][]int{"main.go": {10, 11, 12}},
			perLine:       map[string]map[int]int{"main.go": {10: 3, 11: 1, 12: 5}},
			wantCovered:   3,
			wantCoverable: 3,
		},
		{
			name:          "no changed lines covered",
			changedLines:  map[string][]int{"main.go": {10, 11}},
			perLine:       map[string]map[int]int{"main.go": {10: 0, 11: 0}},
			wantCovered:   0,
			wantCoverable: 2,
		},
		{
			name:          "mixed coverage",
			changedLines:  map[string][]int{"main.go": {10, 11, 12}},
			perLine:       map[string]map[int]int{"main.go": {10: 5, 11: 0, 12: 1}},
			wantCovered:   2,
			wantCoverable: 3,
		},
		{
			name:          "file not input coverage data",
			changedLines:  map[string][]int{"unknown.go": {1, 2, 3}},
			perLine:       map[string]map[int]int{"other.go": {1: 1}},
			wantCovered:   0,
			wantCoverable: 0,
		},
		{
			name:          "changed line not coverable",
			changedLines:  map[string][]int{"main.go": {1, 2, 3}},
			perLine:       map[string]map[int]int{"main.go": {2: 1}},
			wantCovered:   1,
			wantCoverable: 1,
		},
		{
			name:          "nil inputs",
			changedLines:  nil,
			perLine:       nil,
			wantCovered:   0,
			wantCoverable: 0,
		},
		{
			name: "multiple files",
			changedLines: map[string][]int{
				"a.go": {1, 2},
				"b.go": {5},
			},
			perLine: map[string]map[int]int{
				"a.go": {1: 1, 2: 0},
				"b.go": {5: 3},
			},
			wantCovered:   2,
			wantCoverable: 3,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			pc, err := coverage.NewPatchContentFromChanges(tcase.changedLines, tcase.perLine)
			require.NoError(t, err)

			assert.Equal(t, tcase.wantCovered, pc.CoveredLines())
			assert.Equal(t, tcase.wantCoverable, pc.CoverableLines())
		})
	}
}
