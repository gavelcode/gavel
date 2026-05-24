package resources

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/cli/mcp/executor"
)

func fakeCLI(t *testing.T, jsonOutput string) *executor.CLI {
	t.Helper()
	tmpDir := t.TempDir()
	script := filepath.Join(tmpDir, "fake-gavel")
	content := "#!/bin/sh\ncat <<'ENDJSON'\n" + jsonOutput + "\nENDJSON\n"
	require.NoError(t, os.WriteFile(script, []byte(content), 0o755))
	return executor.NewWithBinary(script, tmpDir)
}

func TestReadConfig_FormatsOutput(t *testing.T) {
	configJSON := `{
		"config_path": "/workspace/.gavel/gavel.yaml",
		"gavelspace": "my-monorepo",
		"server": "https://gavel.example.com",
		"projects": [
			{
				"name": "core",
				"pattern": "//core/...",
				"languages": ["go"],
				"quality_gate": {"rules": [{"subtype": "code_quality"}, {"subtype": "coverage"}]}
			}
		]
	}`
	cli := fakeCLI(t, configJSON)

	result, err := readConfig(context.Background(), cli)

	require.NoError(t, err)
	require.Len(t, result.Contents, 1)
	text := result.Contents[0].Text
	assert.Contains(t, text, "gavel.yaml")
	assert.Contains(t, text, "my-monorepo")
	assert.Contains(t, text, "core")
	assert.Contains(t, text, "//core/...")
	assert.Contains(t, text, "go")
	assert.Contains(t, text, "2 rules")
	assert.Contains(t, text, "code_quality")
	assert.Contains(t, text, "coverage")
	assert.Contains(t, text, "https://gavel.example.com")
}

func TestReadConfig_NoGavelspaceOrServer(t *testing.T) {
	configJSON := `{"config_path": "/ws/gavel.yaml", "projects": []}`
	cli := fakeCLI(t, configJSON)

	result, err := readConfig(context.Background(), cli)

	require.NoError(t, err)
	text := result.Contents[0].Text
	assert.Contains(t, text, "Projects: 0")
	assert.NotContains(t, text, "Gavelspace")
	assert.NotContains(t, text, "Server")
}

func TestReadConfig_InvalidJSON_ReturnsRawOutput(t *testing.T) {
	cli := fakeCLI(t, "not json at all")

	result, err := readConfig(context.Background(), cli)

	require.NoError(t, err)
	assert.Contains(t, result.Contents[0].Text, "not json at all")
}

func TestReadProjects_FormatsOutput(t *testing.T) {
	projectsJSON := `{"projects": [
		{"name": "core", "pattern": "//core/...", "languages": ["go"], "quality_gate": {"rules": [{"subtype": "coverage"}]}, "baseline": {"findings_count": 10, "violations_count": 2}},
		{"name": "web", "pattern": "//apps/web/...", "languages": ["typescript"], "quality_gate": {"rules": []}, "baseline": {"findings_count": 0, "violations_count": 0}}
	]}`
	cli := fakeCLI(t, projectsJSON)

	result, err := readProjects(context.Background(), cli)

	require.NoError(t, err)
	text := result.Contents[0].Text
	assert.Contains(t, text, "Projects (2)")
	assert.Contains(t, text, "core")
	assert.Contains(t, text, "//core/...")
	assert.Contains(t, text, "go")
	assert.Contains(t, text, "1 rules")
	assert.Contains(t, text, "web")
	assert.Contains(t, text, "typescript")
}

func TestReadProjects_EmptyList(t *testing.T) {
	cli := fakeCLI(t, `{"projects": []}`)

	result, err := readProjects(context.Background(), cli)

	require.NoError(t, err)
	assert.Contains(t, result.Contents[0].Text, "Projects (0)")
}

func TestReadQualityGate_ReturnsRules(t *testing.T) {
	projectsJSON := `{"projects": [
		{"name": "core", "pattern": "//core/...", "languages": ["go"], "quality_gate": {"rules": [{"subtype": "code_quality"}, {"subtype": "coverage"}, {"subtype": "architecture"}]}, "baseline": {}}
	]}`
	cli := fakeCLI(t, projectsJSON)

	result, err := readQualityGate(context.Background(), cli, "core", "gavel://projects/core/quality-gate")

	require.NoError(t, err)
	text := result.Contents[0].Text
	assert.Contains(t, text, "Quality Gate — core")
	assert.Contains(t, text, "code_quality")
	assert.Contains(t, text, "coverage")
	assert.Contains(t, text, "architecture")
}

func TestReadQualityGate_NoRules(t *testing.T) {
	projectsJSON := `{"projects": [{"name": "empty", "pattern": "//...", "languages": [], "quality_gate": {"rules": []}, "baseline": {}}]}`
	cli := fakeCLI(t, projectsJSON)

	result, err := readQualityGate(context.Background(), cli, "empty", "gavel://projects/empty/quality-gate")

	require.NoError(t, err)
	assert.Contains(t, result.Contents[0].Text, "No quality gate rules")
}

func TestReadQualityGate_ProjectNotFound(t *testing.T) {
	cli := fakeCLI(t, `{"projects": []}`)

	_, err := readQualityGate(context.Background(), cli, "missing", "gavel://projects/missing/quality-gate")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestReadBaseline_WithData(t *testing.T) {
	projectsJSON := `{"projects": [{"name": "core", "pattern": "//core/...", "languages": [], "quality_gate": {"rules": []}, "baseline": {"findings_count": 42, "violations_count": 5}}]}`
	cli := fakeCLI(t, projectsJSON)

	result, err := readBaseline(context.Background(), cli, "core", "gavel://projects/core/baseline")

	require.NoError(t, err)
	text := result.Contents[0].Text
	assert.Contains(t, text, "Baseline — core")
	assert.Contains(t, text, "Findings fingerprints: 42")
	assert.Contains(t, text, "Architecture violations: 5")
}

func TestReadBaseline_NoBaseline(t *testing.T) {
	projectsJSON := `{"projects": [{"name": "new", "pattern": "//...", "languages": [], "quality_gate": {"rules": []}, "baseline": {"findings_count": 0, "violations_count": 0}}]}`
	cli := fakeCLI(t, projectsJSON)

	result, err := readBaseline(context.Background(), cli, "new", "gavel://projects/new/baseline")

	require.NoError(t, err)
	assert.Contains(t, result.Contents[0].Text, "No baseline data")
}

func TestReadBaseline_ProjectNotFound(t *testing.T) {
	cli := fakeCLI(t, `{"projects": []}`)

	_, err := readBaseline(context.Background(), cli, "ghost", "gavel://projects/ghost/baseline")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestReadArchitecture_WithPolicy(t *testing.T) {
	archJSON := `{
		"projects": [],
		"architecture": {
			"layers": [
				{"name": "domain", "patterns": ["core/domain/..."]},
				{"name": "application", "patterns": ["core/application/..."]}
			],
			"deny_rules": [
				{"name": "domain-isolation", "source": "domain", "deny": ["application", "infrastructure"]}
			]
		}
	}`
	cli := fakeCLI(t, archJSON)

	result, err := readArchitecture(context.Background(), cli)

	require.NoError(t, err)
	text := result.Contents[0].Text
	assert.Contains(t, text, "Architecture Policy")
	assert.Contains(t, text, "domain")
	assert.Contains(t, text, "core/domain/...")
	assert.Contains(t, text, "application")
	assert.Contains(t, text, "domain-isolation")
	assert.Contains(t, text, "cannot import")
}

func TestReadArchitecture_NoPolicy(t *testing.T) {
	cli := fakeCLI(t, `{"projects": []}`)

	result, err := readArchitecture(context.Background(), cli)

	require.NoError(t, err)
	assert.Contains(t, result.Contents[0].Text, "No architecture policy")
}

func TestExtractProjectName_Standard(t *testing.T) {
	name := extractProjectName("gavel://projects/core/quality-gate", "/quality-gate")
	assert.Equal(t, "core", name)
}

func TestExtractProjectName_WithHyphens(t *testing.T) {
	name := extractProjectName("gavel://projects/my-project/baseline", "/baseline")
	assert.Equal(t, "my-project", name)
}

func failingCLI(t *testing.T) *executor.CLI {
	t.Helper()
	tmpDir := t.TempDir()
	script := filepath.Join(tmpDir, "failing-gavel")
	content := "#!/bin/sh\necho 'command failed' >&2\nexit 1\n"
	require.NoError(t, os.WriteFile(script, []byte(content), 0o755))
	return executor.NewWithBinary(script, tmpDir)
}

func TestReadArchitecture_CLIError(t *testing.T) {
	cli := failingCLI(t)
	_, err := readArchitecture(context.Background(), cli)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "run gavel projects")
}

func TestReadArchitecture_InvalidJSON(t *testing.T) {
	cli := fakeCLI(t, "not valid json")
	_, err := readArchitecture(context.Background(), cli)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse projects output")
}

func TestReadConfig_CLIError(t *testing.T) {
	cli := failingCLI(t)
	_, err := readConfig(context.Background(), cli)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "run gavel config")
}

func TestReadProjects_CLIError(t *testing.T) {
	cli := failingCLI(t)
	_, err := readProjects(context.Background(), cli)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "run gavel projects")
}

func TestReadQualityGate_CLIError(t *testing.T) {
	cli := failingCLI(t)
	_, err := readQualityGate(context.Background(), cli, "core", "gavel://projects/core/quality-gate")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "run gavel projects")
}

func TestReadBaseline_CLIError(t *testing.T) {
	cli := failingCLI(t)
	_, err := readBaseline(context.Background(), cli, "core", "gavel://projects/core/baseline")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "run gavel projects")
}

func TestFetchProjects_InvalidJSON(t *testing.T) {
	cli := fakeCLI(t, "broken json")
	_, err := fetchProjects(context.Background(), cli)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse projects output")
}

func TestReadQualityGate_SkipsNonMatchingProjects(t *testing.T) {
	projectsJSON := `{"projects": [
		{"name": "alpha", "pattern": "//alpha/...", "languages": [], "quality_gate": {"rules": [{"subtype": "findings"}]}, "baseline": {}},
		{"name": "beta", "pattern": "//beta/...", "languages": [], "quality_gate": {"rules": [{"subtype": "coverage"}]}, "baseline": {}}
	]}`
	cli := fakeCLI(t, projectsJSON)

	result, err := readQualityGate(context.Background(), cli, "beta", "gavel://projects/beta/quality-gate")

	require.NoError(t, err)
	assert.Contains(t, result.Contents[0].Text, "beta")
	assert.Contains(t, result.Contents[0].Text, "coverage")
}

func TestReadBaseline_SkipsNonMatchingProjects(t *testing.T) {
	projectsJSON := `{"projects": [
		{"name": "alpha", "pattern": "//alpha/...", "languages": [], "quality_gate": {"rules": []}, "baseline": {"findings_count": 0, "violations_count": 0}},
		{"name": "beta", "pattern": "//beta/...", "languages": [], "quality_gate": {"rules": []}, "baseline": {"findings_count": 10, "violations_count": 3}}
	]}`
	cli := fakeCLI(t, projectsJSON)

	result, err := readBaseline(context.Background(), cli, "beta", "gavel://projects/beta/baseline")

	require.NoError(t, err)
	assert.Contains(t, result.Contents[0].Text, "beta")
	assert.Contains(t, result.Contents[0].Text, "Findings fingerprints: 10")
}
