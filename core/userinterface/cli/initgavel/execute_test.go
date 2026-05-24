package initgavel

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeInstaller struct {
	installErr error
	workspace  string
	tooling    []string
}

func (f *fakeInstaller) Install(workspace string, tooling []string) (map[string]bool, error) {
	f.workspace = workspace
	f.tooling = tooling
	if f.installErr != nil {
		return nil, f.installErr
	}
	return map[string]bool{
		".bazelrc":     true,
		"MODULE.bazel": true,
	}, nil
}

type fakeCatalog struct{}

func (f *fakeCatalog) Catalog(_ []string) ([]string, []string) {
	return []string{"go_golangci_lint_submission_aspect"}, []string{"golangci_lint_binary"}
}

func validProjects() []Project {
	return []Project{
		{Name: "backend", Pattern: "//server/...", Tooling: []string{"go"}},
	}
}

func TestExecuteCreatesConfig(t *testing.T) {
	workspace := t.TempDir()
	inst := &fakeInstaller{}

	result, err := execute(
		".gavel/gavel.yaml", workspace, "gavel", false,
		validProjects(), Server{}, "", inst, &fakeCatalog{},
	)

	require.NoError(t, err)
	assert.True(t, result.created)
	assert.Equal(t, "gavel", result.name)
	assert.Equal(t, ".gavel/gavel.yaml", result.configPath)
	assert.Len(t, result.projects, 1)
	assert.Equal(t, "backend", result.projects[0].Name)
	assert.Equal(t, workspace, inst.workspace)
	assert.Equal(t, []string{"go"}, inst.tooling)
	assert.Equal(t, []string{"go_golangci_lint_submission_aspect"}, result.aspects)
	assert.Equal(t, []string{"golangci_lint_binary"}, result.binaries)

	configPath := filepath.Join(workspace, ".gavel/gavel.yaml")
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "name: gavel")
	assert.Contains(t, string(data), "pattern: //server/...")
}

func TestExecuteExistingNoForceReturnsNotCreated(t *testing.T) {
	workspace := t.TempDir()
	configDir := filepath.Join(workspace, ".gavel")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "gavel.yaml"), []byte("name: existing"), 0o644))

	inst := &fakeInstaller{}

	result, err := execute(
		".gavel/gavel.yaml", workspace, "gavel", false,
		validProjects(), Server{}, "", inst, &fakeCatalog{},
	)

	require.NoError(t, err)
	assert.False(t, result.created)
	assert.Empty(t, inst.workspace)
}

func TestExecuteExistingWithForceOverwrites(t *testing.T) {
	workspace := t.TempDir()
	configDir := filepath.Join(workspace, ".gavel")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "gavel.yaml"), []byte("name: old"), 0o644))

	inst := &fakeInstaller{}

	result, err := execute(
		".gavel/gavel.yaml", workspace, "gavel", true,
		validProjects(), Server{}, "", inst, &fakeCatalog{},
	)

	require.NoError(t, err)
	assert.True(t, result.created)
	assert.Equal(t, workspace, inst.workspace)
}

func TestExecuteInstallErrorReturnsError(t *testing.T) {
	workspace := t.TempDir()
	inst := &fakeInstaller{installErr: errors.New("permission denied")}

	_, err := execute(
		".gavel/gavel.yaml", workspace, "gavel", false,
		validProjects(), Server{}, "", inst, &fakeCatalog{},
	)

	require.Error(t, err)
	assert.ErrorContains(t, err, "install config")
}

func TestReadFromConfigParsesValidYaml(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gavel.yaml")
	content := `
name: my-monorepo
projects:
  - name: backend
    pattern: //server/...
    tooling: [go, java]
  - name: frontend
    pattern: //web/...
    tooling: [typescript]
server:
  url: https://gavel.example.com
  token: gav_xxx
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	name, projects, server, err := readFromConfig(path)

	require.NoError(t, err)
	assert.Equal(t, "my-monorepo", name)
	require.Len(t, projects, 2)
	assert.Equal(t, "backend", projects[0].Name)
	assert.Equal(t, "//server/...", projects[0].Pattern)
	assert.Equal(t, []string{"go", "java"}, projects[0].Tooling)
	assert.Equal(t, "frontend", projects[1].Name)
	assert.Equal(t, "https://gavel.example.com", server.URL)
	assert.Equal(t, "gav_xxx", server.Token)
}

func TestReadFromConfigRejectsEmptyName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gavel.yaml")
	content := `
projects:
  - name: backend
    pattern: //server/...
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	_, _, _, err := readFromConfig(path)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestReadFromConfigRejectsNoProjects(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gavel.yaml")
	content := `name: my-monorepo`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	_, _, _, err := readFromConfig(path)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one project")
}

func TestReadFromConfigRejectsMissingFile(t *testing.T) {
	_, _, _, err := readFromConfig("/nonexistent/gavel.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "read")
}

func TestReadFromConfigInvalidYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gavel.yaml")
	require.NoError(t, os.WriteFile(path, []byte("{{invalid yaml"), 0o644))

	_, _, _, err := readFromConfig(path)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestResolveAbsConfigPath_Relative(t *testing.T) {
	result := resolveAbsConfigPath("/workspace", ".gavel/gavel.yaml")
	assert.Equal(t, "/workspace/.gavel/gavel.yaml", result)
}

func TestResolveAbsConfigPath_Absolute(t *testing.T) {
	result := resolveAbsConfigPath("/workspace", "/etc/gavel.yaml")
	assert.Equal(t, "/etc/gavel.yaml", result)
}

func TestCopyConfig_Success(t *testing.T) {
	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "gavel.yaml")
	require.NoError(t, os.WriteFile(srcPath, []byte("name: copied"), 0o644))

	dstDir := t.TempDir()
	dstPath := filepath.Join(dstDir, ".gavel", "gavel.yaml")

	err := copyConfig(srcPath, dstPath)

	require.NoError(t, err)
	data, err := os.ReadFile(dstPath)
	require.NoError(t, err)
	assert.Equal(t, "name: copied", string(data))
}

func TestCopyConfig_ReadError(t *testing.T) {
	dstPath := filepath.Join(t.TempDir(), "out.yaml")
	err := copyConfig("/nonexistent/gavel.yaml", dstPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read source config")
}

func TestCopyConfig_MkdirError(t *testing.T) {
	srcPath := filepath.Join(t.TempDir(), "src.yaml")
	require.NoError(t, os.WriteFile(srcPath, []byte("ok"), 0o644))

	err := copyConfig(srcPath, "/dev/null/subdir/out.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create config dir")
}

func TestWriteConfig_Success(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".gavel", "gavel.yaml")

	err := writeConfig(path, "test", validProjects(), Server{})

	require.NoError(t, err)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "name: test")
}

func TestWriteConfig_MkdirError(t *testing.T) {
	err := writeConfig("/dev/null/sub/gavel.yaml", "test", validProjects(), Server{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create config dir")
}

func TestBuildConfigDTO_WithServer(t *testing.T) {
	dto := buildConfigDTO("proj", validProjects(), Server{URL: "https://x.com", Token: "tok"})
	assert.Equal(t, "proj", dto.Name)
	assert.Equal(t, "https://x.com", dto.Server.URL)
	assert.Equal(t, "tok", dto.Server.Token)
	require.Len(t, dto.Projects, 1)
	require.NotNil(t, dto.Projects[0].Gate)
}

func TestDefaultQualityGate(t *testing.T) {
	gate := defaultQualityGate()
	require.NotNil(t, gate.Findings)
	require.NotNil(t, gate.Coverage)
	require.NotNil(t, gate.Violations)
	assert.Equal(t, 0, gate.Findings.MaxError)
	assert.Equal(t, defaultCoverageMin, gate.Coverage.Min)
	assert.Equal(t, 0, gate.Violations.Max)
}

func TestExtractTooling_Deduplication(t *testing.T) {
	projects := []Project{
		{Name: "a", Tooling: []string{"go", "java"}},
		{Name: "b", Tooling: []string{"java", "python"}},
	}
	result := extractTooling(projects)
	assert.Equal(t, []string{"go", "java", "python"}, result)
}

func TestExecuteWithSourceFile(t *testing.T) {
	workspace := t.TempDir()
	srcFile := filepath.Join(t.TempDir(), "gavel.yaml")
	require.NoError(t, os.WriteFile(srcFile, []byte("name: from-source"), 0o644))

	inst := &fakeInstaller{}
	result, err := execute(
		".gavel/gavel.yaml", workspace, "proj", true,
		validProjects(), Server{}, srcFile, inst, &fakeCatalog{},
	)

	require.NoError(t, err)
	assert.True(t, result.created)
	data, err := os.ReadFile(filepath.Join(workspace, ".gavel/gavel.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "name: from-source", string(data))
}

func TestExecuteCopyConfigError(t *testing.T) {
	workspace := t.TempDir()
	inst := &fakeInstaller{}

	_, err := execute(
		".gavel/gavel.yaml", workspace, "proj", true,
		validProjects(), Server{}, "/nonexistent/config.yaml", inst, &fakeCatalog{},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "copy config")
}

func TestExecuteWriteConfigError(t *testing.T) {
	inst := &fakeInstaller{}

	_, err := execute(
		"/dev/null/sub/gavel.yaml", "/dev/null", "proj", false,
		validProjects(), Server{}, "", inst, &fakeCatalog{},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "save config")
}

func TestExecuteArchConfigError(t *testing.T) {
	workspace := t.TempDir()
	archPath := filepath.Join(workspace, ".gavel")
	require.NoError(t, os.MkdirAll(archPath, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(archPath, "architecture.yml"), nil, 0o000))
	require.NoError(t, os.Chmod(filepath.Join(archPath, "architecture.yml"), 0o000))

	inst := &fakeInstaller{}

	_, err := execute(
		".gavel/gavel.yaml", workspace, "proj", false,
		validProjects(), Server{}, "", inst, &fakeCatalog{},
	)

	t.Cleanup(func() { require.NoError(t, os.Chmod(filepath.Join(archPath, "architecture.yml"), 0o644)) })

	if err != nil {
		assert.Contains(t, err.Error(), "save config")
	}
}

func TestExecuteFromConfigRunsInstaller(t *testing.T) {
	workspace := t.TempDir()
	configFile := filepath.Join(t.TempDir(), "gavel.yaml")
	content := `
name: my-project
projects:
  - name: api
    pattern: //api/...
    tooling: [go]
`
	require.NoError(t, os.WriteFile(configFile, []byte(content), 0o644))

	name, projects, server, err := readFromConfig(configFile)
	require.NoError(t, err)

	inst := &fakeInstaller{}
	result, err := execute(
		".gavel/gavel.yaml", workspace, name, true,
		projects, server, "", inst, &fakeCatalog{},
	)

	require.NoError(t, err)
	assert.True(t, result.created)
	assert.Equal(t, "my-project", result.name)
	assert.Equal(t, workspace, inst.workspace)
	assert.Equal(t, []string{"go"}, inst.tooling)
}
