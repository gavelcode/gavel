package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunFindings_FlattensAllProjects(t *testing.T) {
	findingsJSON := `{"projects":[
		{"name":"core","findings":[{"tool":"golangci-lint","rule_id":"errcheck","severity":"error","file_path":"core/a.go","line":5,"message":"unchecked error","fingerprint":"fp1"}]},
		{"name":"cli","findings":[{"tool":"golangci-lint","rule_id":"varnamelen","severity":"warning","file_path":"cli/b.go","line":10,"message":"name too short","fingerprint":"fp2"}]}
	]}`
	cli := fakeCLI(t, findingsJSON)

	result, err := runFindings(context.Background(), cli, FindingsInput{})

	require.NoError(t, err)
	assert.Contains(t, result, "core/a.go")
	assert.Contains(t, result, "cli/b.go")
	assert.Contains(t, result, "# Findings (2)")
}

func TestRunFindings_ByRuleSummary(t *testing.T) {
	findingsJSON := `{"projects":[{"name":"core","findings":[
		{"tool":"golangci-lint","rule_id":"varnamelen","severity":"warning","file_path":"a.go","line":1,"message":"x","fingerprint":"f1"},
		{"tool":"golangci-lint","rule_id":"varnamelen","severity":"warning","file_path":"b.go","line":2,"message":"y","fingerprint":"f2"},
		{"tool":"golangci-lint","rule_id":"errcheck","severity":"error","file_path":"c.go","line":3,"message":"z","fingerprint":"f3"}
	]}]}`
	cli := fakeCLI(t, findingsJSON)

	result, err := runFindings(context.Background(), cli, FindingsInput{})

	require.NoError(t, err)
	assert.Contains(t, result, "By rule:")
	assert.Contains(t, result, "2 varnamelen")
	assert.Contains(t, result, "1 errcheck")
}

func TestRunFindings_RuleFilter(t *testing.T) {
	findingsJSON := `{"projects":[{"name":"core","findings":[
		{"tool":"golangci-lint","rule_id":"varnamelen","severity":"warning","file_path":"a.go","line":1,"message":"short name","fingerprint":"f1"},
		{"tool":"golangci-lint","rule_id":"errcheck","severity":"error","file_path":"b.go","line":2,"message":"unchecked","fingerprint":"f2"}
	]}]}`
	cli := fakeCLI(t, findingsJSON)

	result, err := runFindings(context.Background(), cli, FindingsInput{Rule: "varnamelen"})

	require.NoError(t, err)
	assert.Contains(t, result, "a.go")
	assert.NotContains(t, result, "b.go")
	assert.Contains(t, result, "# Findings (1)")
}

func TestRunFindings_SeverityFilter(t *testing.T) {
	findingsJSON := `{"projects":[{"name":"core","findings":[
		{"tool":"golangci-lint","rule_id":"varnamelen","severity":"warning","file_path":"a.go","line":1,"message":"short name","fingerprint":"f1"},
		{"tool":"golangci-lint","rule_id":"errcheck","severity":"error","file_path":"b.go","line":2,"message":"unchecked","fingerprint":"f2"}
	]}]}`
	cli := fakeCLI(t, findingsJSON)

	result, err := runFindings(context.Background(), cli, FindingsInput{Severity: "error"})

	require.NoError(t, err)
	assert.Contains(t, result, "b.go")
	assert.NotContains(t, result, "a.go")
}

func TestRunFindings_NewFindingsSortFirst(t *testing.T) {
	findingsJSON := `{"projects":[{"name":"core","findings":[
		{"tool":"golangci-lint","rule_id":"errcheck","severity":"error","file_path":"aaa.go","line":1,"message":"baselined","fingerprint":"f1"},
		{"tool":"golangci-lint","rule_id":"varnamelen","severity":"warning","file_path":"zzz.go","line":99,"message":"brand new","fingerprint":"f2","status":"new"}
	]}]}`
	cli := fakeCLI(t, findingsJSON)

	result, err := runFindings(context.Background(), cli, FindingsInput{})

	require.NoError(t, err)
	assert.Contains(t, result, "NEW")
	assert.Less(t, strings.Index(result, "zzz.go"), strings.Index(result, "aaa.go"))
}

func TestRunFindings_LimitTruncatesListButNotCounts(t *testing.T) {
	findingsJSON := `{"projects":[{"name":"core","findings":[
		{"tool":"golangci-lint","rule_id":"errcheck","severity":"error","file_path":"a.go","line":1,"message":"one","fingerprint":"f1"},
		{"tool":"golangci-lint","rule_id":"errcheck","severity":"error","file_path":"b.go","line":2,"message":"two","fingerprint":"f2"},
		{"tool":"golangci-lint","rule_id":"errcheck","severity":"error","file_path":"c.go","line":3,"message":"three","fingerprint":"f3"}
	]}]}`
	cli := fakeCLI(t, findingsJSON)

	result, err := runFindings(context.Background(), cli, FindingsInput{Limit: 2})

	require.NoError(t, err)
	assert.Contains(t, result, "# Findings (3)")
	assert.Contains(t, result, "3 errcheck")
	assert.Contains(t, result, "Showing 2 of 3")
}

func TestRunFindings_NoFindings(t *testing.T) {
	cli := fakeCLI(t, `{"projects":[{"name":"core","findings":[]}]}`)

	result, err := runFindings(context.Background(), cli, FindingsInput{})

	require.NoError(t, err)
	assert.Contains(t, result, "No findings")
}

func TestRunFindings_WithProjectFilter(t *testing.T) {
	findingsJSON := `{"projects":[{"name":"web","findings":[{"tool":"eslint","rule_id":"no-unused-vars","severity":"warning","file_path":"app.ts","line":5,"message":"unused","fingerprint":"f1"}]}]}`
	cli := fakeCLI(t, findingsJSON)

	result, err := runFindings(context.Background(), cli, FindingsInput{Project: "web"})

	require.NoError(t, err)
	assert.Contains(t, result, "app.ts")
}

func TestRunFindings_CLIError(t *testing.T) {
	cli := failingCLI(t)
	_, err := runFindings(context.Background(), cli, FindingsInput{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "execute gavel judge")
}

func TestRunFindings_ParseError(t *testing.T) {
	cli := fakeCLI(t, "not json")
	_, err := runFindings(context.Background(), cli, FindingsInput{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse judge output")
}
