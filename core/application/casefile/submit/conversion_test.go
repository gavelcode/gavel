package submit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
)

func TestToFileCoverageEntriesEmptyReturnsNil(t *testing.T) {
	assert.Nil(t, toFileCoverageEntries(nil))
}

func TestToFileCoverageEntriesConvertsValidDTOs(t *testing.T) {
	dtos := []evidencedto.FileCoverage{
		{FilePath: "pkg/a.go", Covered: []int{1, 2}, Uncovered: []int{3}},
		{FilePath: "pkg/b.go", Covered: []int{1}, Uncovered: []int{2, 3}},
	}

	got := toFileCoverageEntries(dtos)

	assert.Len(t, got, 2)
}

func TestToFileCoverageEntriesSkipsInvalidEntries(t *testing.T) {
	dtos := []evidencedto.FileCoverage{
		{FilePath: "", Covered: []int{1}, Uncovered: []int{2}},
		{FilePath: "valid.go", Covered: []int{1}, Uncovered: []int{2}},
	}

	got := toFileCoverageEntries(dtos)

	assert.Len(t, got, 1)
	assert.Equal(t, "valid.go", got[0].FilePath())
}
