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
