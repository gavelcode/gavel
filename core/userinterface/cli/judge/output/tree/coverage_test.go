package tree_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/cli/judge/output/tree"
)

func TestBuildCoverageTreeEmpty(t *testing.T) {
	root := tree.BuildCoverageTree(nil)

	assert.Equal(t, "", root.Path)
	assert.Equal(t, 0, root.CoveredLines)
	assert.Equal(t, 0, root.TotalLines)
	assert.Empty(t, root.Children)
	assert.Empty(t, root.Files)
}

func TestBuildCoverageTreeSingleFile(t *testing.T) {
	files := []tree.FileCoverage{
		{FilePath: "core/domain/model.go", CoveredLines: 80, TotalLines: 100},
	}

	root := tree.BuildCoverageTree(files)

	assert.Equal(t, 80, root.CoveredLines)
	assert.Equal(t, 100, root.TotalLines)
	assert.InDelta(t, 80.0, root.Percent, 0.1)

	require.Len(t, root.Children, 1)
	core := root.Children[0]
	assert.Equal(t, "core", core.Path)

	require.Len(t, core.Children, 1)
	domain := core.Children[0]
	assert.Equal(t, "core/domain", domain.Path)

	require.Len(t, domain.Files, 1)
	assert.Equal(t, "model.go", domain.Files[0].Name)
	assert.Equal(t, 80, domain.Files[0].CoveredLines)
}

func TestBuildCoverageTreeAggregatesDirectory(t *testing.T) {
	files := []tree.FileCoverage{
		{FilePath: "pkg/a.go", CoveredLines: 50, TotalLines: 100},
		{FilePath: "pkg/b.go", CoveredLines: 30, TotalLines: 50},
	}

	root := tree.BuildCoverageTree(files)

	require.Len(t, root.Children, 1)
	pkg := root.Children[0]
	assert.Equal(t, "pkg", pkg.Path)
	assert.Equal(t, 80, pkg.CoveredLines)
	assert.Equal(t, 150, pkg.TotalLines)
	assert.InDelta(t, 53.3, pkg.Percent, 0.1)
	assert.Len(t, pkg.Files, 2)
}

func TestBuildCoverageTreeNestedDirectories(t *testing.T) {
	files := []tree.FileCoverage{
		{FilePath: "a/b/c/x.go", CoveredLines: 10, TotalLines: 20},
		{FilePath: "a/b/d/y.go", CoveredLines: 5, TotalLines: 10},
		{FilePath: "a/e/z.go", CoveredLines: 15, TotalLines: 30},
	}

	root := tree.BuildCoverageTree(files)

	assert.Equal(t, 30, root.CoveredLines)
	assert.Equal(t, 60, root.TotalLines)

	require.Len(t, root.Children, 1)
	a := root.Children[0]
	assert.Equal(t, "a", a.Path)
	assert.Equal(t, 30, a.CoveredLines)
	assert.Equal(t, 60, a.TotalLines)

	require.Len(t, a.Children, 2)
}

func TestBuildCoverageTreeSortsChildren(t *testing.T) {
	files := []tree.FileCoverage{
		{FilePath: "z/f.go", CoveredLines: 1, TotalLines: 1},
		{FilePath: "a/f.go", CoveredLines: 1, TotalLines: 1},
		{FilePath: "m/f.go", CoveredLines: 1, TotalLines: 1},
	}

	root := tree.BuildCoverageTree(files)

	require.Len(t, root.Children, 3)
	assert.Equal(t, "a", root.Children[0].Path)
	assert.Equal(t, "m", root.Children[1].Path)
	assert.Equal(t, "z", root.Children[2].Path)
}

func TestBuildCoverageTreeSortsFiles(t *testing.T) {
	files := []tree.FileCoverage{
		{FilePath: "pkg/z.go", CoveredLines: 1, TotalLines: 1},
		{FilePath: "pkg/a.go", CoveredLines: 1, TotalLines: 1},
	}

	root := tree.BuildCoverageTree(files)

	require.Len(t, root.Children, 1)
	require.Len(t, root.Children[0].Files, 2)
	assert.Equal(t, "a.go", root.Children[0].Files[0].Name)
	assert.Equal(t, "z.go", root.Children[0].Files[1].Name)
}

func TestBuildCoverageTreeZeroTotalLines(t *testing.T) {
	files := []tree.FileCoverage{
		{FilePath: "pkg/empty.go", CoveredLines: 0, TotalLines: 0},
	}

	root := tree.BuildCoverageTree(files)

	require.Len(t, root.Children, 1)
	require.Len(t, root.Children[0].Files, 1)
	assert.Equal(t, 0.0, root.Children[0].Files[0].Percent)
}
