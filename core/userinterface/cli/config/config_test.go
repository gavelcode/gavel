package config_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/gavelspace/loadgavelspace"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	gavelspacemodel "github.com/usegavel/gavel/core/domain/gavelspace/model"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
	"github.com/usegavel/gavel/core/userinterface/cli/config"
)

var testUUID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

type fakeFinder struct {
	gavelspace gavelspacemodel.Gavelspace
	projects   []projectmodel.Project
	err        error
}

func (f fakeFinder) LoadFromConfig(_ string) (gavelspacemodel.Gavelspace, []projectmodel.Project, error) {
	return f.gavelspace, f.projects, f.err
}

func workspaceOK() (string, error)   { return "/tmp/ws", nil }
func workspaceFail() (string, error) { return "", errors.New("no workspace") }

func newGavelspace(t *testing.T, name string) gavelspacemodel.Gavelspace {
	t.Helper()
	gs, err := gavelspacemodel.NewGavelspace(tenant.LocalTenantID, name)
	require.NoError(t, err)
	return gs
}

func newProject(t *testing.T, name, pattern string) projectmodel.Project {
	t.Helper()
	p, err := projectmodel.NewProject(name+"-key", name, pattern)
	require.NoError(t, err)
	return p
}

func newProjectWithGate(t *testing.T, name, pattern string) projectmodel.Project {
	t.Helper()
	strategy, err := qualitygate.NewCountBySeverity(0, 0, 0)
	require.NoError(t, err)
	rule, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, strategy)
	require.NoError(t, err)
	gate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)
	project, err := projectmodel.ReconstituteProject(
		projectmodel.NewProjectID(testUUID),
		name+"-key", name, pattern, "main",
		nil, gate, nil, nil,
	)
	require.NoError(t, err)
	return project
}

func TestConfigOutputsJSON(t *testing.T) {
	gs := newGavelspace(t, "myrepo")
	proj := newProjectWithGate(t, "core", "//core/...")
	finder := fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}}
	handler := loadgavelspace.NewHandler(finder)

	cmd := config.NewCommand(workspaceOK, handler)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--config", "/tmp/gavel.yaml"})

	err := cmd.Execute()

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "/tmp/gavel.yaml", out["config_path"])
	assert.Equal(t, "myrepo", out["gavelspace"])
	projects := out["projects"].([]any)
	require.Len(t, projects, 1)
	p := projects[0].(map[string]any)
	assert.Equal(t, "core", p["name"])
	assert.Equal(t, "//core/...", p["pattern"])
	gate := p["quality_gate"].(map[string]any)
	rules := gate["rules"].([]any)
	require.Len(t, rules, 1)
	assert.Equal(t, "code_quality", rules[0].(map[string]any)["subtype"])
}

func TestConfigWorkspaceResolverError(t *testing.T) {
	handler := loadgavelspace.NewHandler(fakeFinder{})
	cmd := config.NewCommand(workspaceFail, handler)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no workspace")
}

func TestConfigHandlerError(t *testing.T) {
	finder := fakeFinder{err: errors.New("bad config")}
	handler := loadgavelspace.NewHandler(finder)

	cmd := config.NewCommand(workspaceOK, handler)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--config", "/tmp/gavel.yaml"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load config")
}

func TestConfigResolvesConfigPathFromGavelDir(t *testing.T) {
	dir := t.TempDir()
	gavelDir := filepath.Join(dir, ".gavel")
	require.NoError(t, os.MkdirAll(gavelDir, 0o755))
	configFile := filepath.Join(gavelDir, "gavel.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte("name: test"), 0o644))

	gs := newGavelspace(t, "test")
	finder := fakeFinder{gavelspace: gs}
	handler := loadgavelspace.NewHandler(finder)

	resolver := func() (string, error) { return dir, nil }
	cmd := config.NewCommand(resolver, handler)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, configFile, out["config_path"])
}

func TestConfigResolvesConfigPathFromRoot(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "gavel.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte("name: test"), 0o644))

	gs := newGavelspace(t, "test")
	finder := fakeFinder{gavelspace: gs}
	handler := loadgavelspace.NewHandler(finder)

	resolver := func() (string, error) { return dir, nil }
	cmd := config.NewCommand(resolver, handler)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, configFile, out["config_path"])
}

func TestConfigFallsBackToGavelDirPath(t *testing.T) {
	dir := t.TempDir()

	finder := fakeFinder{err: errors.New("not found")}
	handler := loadgavelspace.NewHandler(finder)

	resolver := func() (string, error) { return dir, nil }
	cmd := config.NewCommand(resolver, handler)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load config")
}

func TestConfigOmitsEmptyGavelspaceAndServer(t *testing.T) {
	finder := fakeFinder{}
	handler := loadgavelspace.NewHandler(finder)

	cmd := config.NewCommand(workspaceOK, handler)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--config", "/tmp/gavel.yaml"})

	err := cmd.Execute()

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	_, hasGavelspace := out["gavelspace"]
	_, hasServer := out["server"]
	assert.False(t, hasGavelspace)
	assert.False(t, hasServer)
}

func TestConfigIncludesServerURL(t *testing.T) {
	gs := newGavelspace(t, "myrepo")
	gs.SetServerConfig(gavelspacemodel.NewServerConfig("https://gavel.example.com", "tok"))

	finder := fakeFinder{gavelspace: gs, projects: []projectmodel.Project{newProject(t, "core", "//core/...")}}
	handler := loadgavelspace.NewHandler(finder)

	cmd := config.NewCommand(workspaceOK, handler)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--config", "/tmp/gavel.yaml"})

	err := cmd.Execute()

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "https://gavel.example.com", out["server"])
}

func TestRegisterFlagsBindsConfigFlag(t *testing.T) {
	cmd := &cobra.Command{}
	opts := &config.Options{}

	config.RegisterFlags(cmd, opts)

	assert.NotNil(t, cmd.Flags().Lookup("config"))
}
