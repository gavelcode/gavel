package tools

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const coverageJSON = `{"projects":[{"name":"core","verdict":"fail","coverage_percent":50.0,
	"rulings":[{"subtype":"coverage","passed":false,"detail":"50.0% coverage (min 90.0%)"}],
	"coverage_by_file":[
		{"file_path":"a.go","covered_lines":2,"total_lines":4,"percent":50.0,"covered":[1,2],"uncovered":[3,4]},
		{"file_path":"b.go","covered_lines":3,"total_lines":3,"percent":100.0,"covered":[1,2,3],"uncovered":[]}
	]}]}`

func TestFormatCoverage_NoFilter_ShowsPerFileTable(t *testing.T) {
	var resp judgeResponse
	require.NoError(t, json.Unmarshal([]byte(coverageJSON), &resp))

	out := formatCoverage(resp, nil)

	assert.Contains(t, out, "Coverage: 50.0%")
	assert.Contains(t, out, "a.go")
	assert.Contains(t, out, "50.0% (2/4)")
	assert.Contains(t, out, "b.go")
	assert.Contains(t, out, "100.0% (3/3)")
}

func TestFormatCoverage_WithFilter_ShowsOnlyMatchAndUncovered(t *testing.T) {
	var resp judgeResponse
	require.NoError(t, json.Unmarshal([]byte(coverageJSON), &resp))

	out := formatCoverage(resp, []string{"a.go"})

	assert.Contains(t, out, "a.go")
	assert.Contains(t, out, "50.0% (2/4)")
	assert.Contains(t, out, "uncovered: 3-4")
	assert.False(t, strings.Contains(out, "b.go"), "filtered-out file must not appear")
}

const coverageTreeJSON = `{"projects":[{"name":"core","verdict":"pass","coverage_percent":75.0,
	"rulings":[{"subtype":"coverage","passed":true,"detail":"75.0% coverage (min 70.0%)"}],
	"coverage_tree":{
		"path":"","covered_lines":9,"total_lines":12,"percent":75.0,
		"children":[
			{"path":"pkg","covered_lines":9,"total_lines":12,"percent":75.0,
			 "children":[
				{"path":"pkg/sub","covered_lines":5,"total_lines":8,"percent":62.5,
				 "files":[
					{"name":"a.go","covered_lines":3,"total_lines":4,"percent":75.0},
					{"name":"b.go","covered_lines":2,"total_lines":4,"percent":50.0}
				 ]},
				{"path":"pkg/other","covered_lines":4,"total_lines":4,"percent":100.0,
				 "files":[
					{"name":"c.go","covered_lines":4,"total_lines":4,"percent":100.0}
				 ]}
			 ]}
		]
	}}]}`

func TestFormatCoverage_WithTree_ShowsDirectoryHierarchy(t *testing.T) {
	var resp judgeResponse
	require.NoError(t, json.Unmarshal([]byte(coverageTreeJSON), &resp))

	out := formatCoverage(resp, nil)

	assert.Contains(t, out, "Coverage: 75.0%")
	assert.Contains(t, out, "pkg/sub")
	assert.Contains(t, out, "62.5%")
	assert.Contains(t, out, "pkg/other")
	assert.Contains(t, out, "100.0%")
	assert.Contains(t, out, "a.go")
	assert.Contains(t, out, "b.go")
	assert.Contains(t, out, "c.go")
}

func TestFormatCoverage_WithTree_FilterStillUsesFileBreakdown(t *testing.T) {
	var resp judgeResponse
	require.NoError(t, json.Unmarshal([]byte(coverageJSON), &resp))
	resp.Projects[0].CoverageTree = &judgeCoverageNode{
		Path: "", CoveredLines: 5, TotalLines: 7, Percent: 71.4,
	}

	out := formatCoverage(resp, []string{"a.go"})

	assert.Contains(t, out, "a.go")
	assert.Contains(t, out, "uncovered: 3-4")
}

const packagesJSON = `{"projects":[{"name":"core","verdict":"fail","coverage_percent":60.0,
	"coverage_by_file":[
		{"file_path":"core/a/x.go","covered_lines":1,"total_lines":2,"percent":50.0,"covered":[1],"uncovered":[2]},
		{"file_path":"core/a/sub/y.go","covered_lines":2,"total_lines":2,"percent":100.0,"covered":[1,2],"uncovered":[]},
		{"file_path":"core/b/z.go","covered_lines":1,"total_lines":2,"percent":50.0,"covered":[1],"uncovered":[2]}
	]}]}`

func TestValidateCoverageInput_PackagesAndFiles_Errors(t *testing.T) {
	err := validateCoverageInput(CoverageInput{
		Files:    []string{"core/a/x.go"},
		Packages: []string{"core/a/..."},
	})

	require.Error(t, err)
}

func TestValidateCoverageInput_OnlyOne_OK(t *testing.T) {
	require.NoError(t, validateCoverageInput(CoverageInput{Files: []string{"core/a/x.go"}}))
	require.NoError(t, validateCoverageInput(CoverageInput{Packages: []string{"core/a/..."}}))
	require.NoError(t, validateCoverageInput(CoverageInput{}))
}

func TestResolvePackages_Recursive_IncludesSubpackages(t *testing.T) {
	var resp judgeResponse
	require.NoError(t, json.Unmarshal([]byte(packagesJSON), &resp))

	files, err := resolvePackages(resp, []string{"core/a/..."})

	require.NoError(t, err)
	assert.Equal(t, []string{"core/a/sub/y.go", "core/a/x.go"}, files)
}

func TestResolvePackages_ExactPackage_ExcludesSubpackages(t *testing.T) {
	var resp judgeResponse
	require.NoError(t, json.Unmarshal([]byte(packagesJSON), &resp))

	files, err := resolvePackages(resp, []string{"core/a"})

	require.NoError(t, err)
	assert.Equal(t, []string{"core/a/x.go"}, files)
}

func TestResolvePackages_NoMatch_Errors(t *testing.T) {
	var resp judgeResponse
	require.NoError(t, json.Unmarshal([]byte(packagesJSON), &resp))

	_, err := resolvePackages(resp, []string{"core/nope/..."})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "core/nope/...")
}

func TestResolvePackages_Deduplicates(t *testing.T) {
	var resp judgeResponse
	require.NoError(t, json.Unmarshal([]byte(packagesJSON), &resp))

	files, err := resolvePackages(resp, []string{"core/a/...", "core/a"})

	require.NoError(t, err)
	assert.Equal(t, []string{"core/a/sub/y.go", "core/a/x.go"}, files)
}
