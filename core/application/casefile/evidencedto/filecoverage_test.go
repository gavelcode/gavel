package evidencedto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
)

func TestFileCoverageFromPerLine_EmptyMap(t *testing.T) {
	got := evidencedto.FileCoverageFromPerLine(nil)

	assert.Empty(t, got)
}

func TestFileCoverageFromPerLine_SplitsCoveredAndUncovered(t *testing.T) {
	perLine := map[string]map[int]int{
		"main.go": {3: 0, 1: 2, 2: 1},
	}

	got := evidencedto.FileCoverageFromPerLine(perLine)

	want := []evidencedto.FileCoverage{
		{FilePath: "main.go", Covered: []int{1, 2}, Uncovered: []int{3}},
	}
	assert.Equal(t, want, got)
}

func TestFileCoverageFromPerLine_SortsFilesByPath(t *testing.T) {
	perLine := map[string]map[int]int{
		"b.go": {1: 1},
		"a.go": {1: 0},
		"c.go": {1: 1},
	}

	got := evidencedto.FileCoverageFromPerLine(perLine)

	paths := []string{got[0].FilePath, got[1].FilePath, got[2].FilePath}
	assert.Equal(t, []string{"a.go", "b.go", "c.go"}, paths)
}

func TestFileCoverageFromPerLine_SortsLineNumbers(t *testing.T) {
	perLine := map[string]map[int]int{
		"main.go": {10: 1, 2: 1, 7: 0, 1: 0, 5: 1},
	}

	got := evidencedto.FileCoverageFromPerLine(perLine)

	assert.Equal(t, []int{2, 5, 10}, got[0].Covered)
	assert.Equal(t, []int{1, 7}, got[0].Uncovered)
}
