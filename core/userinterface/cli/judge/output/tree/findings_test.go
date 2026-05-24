package tree_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/cli/judge/output/tree"
)

func TestBuildFindingsTreeEmpty(t *testing.T) {
	root := tree.BuildFindingsTree(nil)

	assert.Equal(t, "", root.Path)
	assert.Equal(t, 0, root.Count)
	assert.Empty(t, root.BySeverity)
	assert.Empty(t, root.Children)
	assert.Empty(t, root.Files)
}

func TestBuildFindingsTreeSingleFile(t *testing.T) {
	findings := []tree.FindingInput{
		{FilePath: "core/domain/model.go", Severity: "error"},
		{FilePath: "core/domain/model.go", Severity: "error"},
		{FilePath: "core/domain/model.go", Severity: "warning"},
	}

	root := tree.BuildFindingsTree(findings)

	assert.Equal(t, 3, root.Count)
	assert.Equal(t, 2, root.BySeverity["error"])
	assert.Equal(t, 1, root.BySeverity["warning"])

	require.Len(t, root.Children, 1)
	core := root.Children[0]
	assert.Equal(t, "core", core.Path)
	assert.Equal(t, 3, core.Count)

	require.Len(t, core.Children, 1)
	domain := core.Children[0]
	assert.Equal(t, "core/domain", domain.Path)

	require.Len(t, domain.Files, 1)
	assert.Equal(t, "model.go", domain.Files[0].Name)
	assert.Equal(t, 3, domain.Files[0].Count)
	assert.Equal(t, 2, domain.Files[0].BySeverity["error"])
}

func TestBuildFindingsTreeAggregatesDirectory(t *testing.T) {
	findings := []tree.FindingInput{
		{FilePath: "pkg/a.go", Severity: "error"},
		{FilePath: "pkg/b.go", Severity: "warning"},
		{FilePath: "pkg/b.go", Severity: "error"},
	}

	root := tree.BuildFindingsTree(findings)

	require.Len(t, root.Children, 1)
	pkg := root.Children[0]
	assert.Equal(t, 3, pkg.Count)
	assert.Equal(t, 2, pkg.BySeverity["error"])
	assert.Equal(t, 1, pkg.BySeverity["warning"])
	assert.Len(t, pkg.Files, 2)
}

func TestBuildFindingsTreeSortsChildren(t *testing.T) {
	findings := []tree.FindingInput{
		{FilePath: "z/f.go", Severity: "error"},
		{FilePath: "a/f.go", Severity: "error"},
	}

	root := tree.BuildFindingsTree(findings)

	require.Len(t, root.Children, 2)
	assert.Equal(t, "a", root.Children[0].Path)
	assert.Equal(t, "z", root.Children[1].Path)
}

func TestBuildFindingsTreeSortsFiles(t *testing.T) {
	findings := []tree.FindingInput{
		{FilePath: "pkg/z.go", Severity: "error"},
		{FilePath: "pkg/a.go", Severity: "error"},
	}

	root := tree.BuildFindingsTree(findings)

	require.Len(t, root.Children, 1)
	require.Len(t, root.Children[0].Files, 2)
	assert.Equal(t, "a.go", root.Children[0].Files[0].Name)
	assert.Equal(t, "z.go", root.Children[0].Files[1].Name)
}

func TestBuildFindingsTreeMultipleSeverities(t *testing.T) {
	findings := []tree.FindingInput{
		{FilePath: "a/x.go", Severity: "error"},
		{FilePath: "a/x.go", Severity: "note"},
		{FilePath: "b/y.go", Severity: "warning"},
	}

	root := tree.BuildFindingsTree(findings)

	assert.Equal(t, 3, root.Count)
	assert.Equal(t, 1, root.BySeverity["error"])
	assert.Equal(t, 1, root.BySeverity["warning"])
	assert.Equal(t, 1, root.BySeverity["note"])
}
