package projects_test

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
	"github.com/usegavel/gavel/core/domain/project/model/archpolicy"
	"github.com/usegavel/gavel/core/domain/project/model/qualitygate"
	"github.com/usegavel/gavel/core/userinterface/cli/projects"
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

func newProjectWithGate(t *testing.T, name, pattern string) projectmodel.Project {
	t.Helper()
	strategy, err := qualitygate.NewCountBySeverity(0, 0, 0)
	require.NoError(t, err)
	rule, err := qualitygate.NewRule(evidence.SubtypeCodeQuality, strategy)
	require.NoError(t, err)
	gate, err := qualitygate.NewGate([]qualitygate.Rule{rule})
	require.NoError(t, err)
	proj, err := projectmodel.ReconstituteProject(
		projectmodel.NewProjectID(testUUID),
		name+"-key", name, pattern, "main",
		nil, gate, nil, nil,
	)
	require.NoError(t, err)
	return proj
}

func newProjectWithArchPolicy(t *testing.T, name, pattern string) projectmodel.Project {
	t.Helper()
	domainLayer, err := archpolicy.NewLayer("domain", []string{"**/domain/**"})
	require.NoError(t, err)
	infraLayer, err := archpolicy.NewLayer("infrastructure", []string{"**/infrastructure/**"})
	require.NoError(t, err)
	denyRule, err := archpolicy.NewDenyRule("domain-isolation", "domain", []string{"infrastructure"})
	require.NoError(t, err)
	policy, err := archpolicy.NewPolicy([]archpolicy.Layer{domainLayer, infraLayer}, []archpolicy.DenyRule{denyRule}, false)
	require.NoError(t, err)
	proj, err := projectmodel.ReconstituteProject(
		projectmodel.NewProjectID(testUUID),
		name+"-key", name, pattern, "main",
		nil, qualitygate.Gate{}, &policy, nil,
	)
	require.NoError(t, err)
	return proj
}

func TestProjectsOutputsJSON(t *testing.T) {
	gs := newGavelspace(t, "myrepo")
	proj := newProjectWithGate(t, "core", "//core/...")
	finder := fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}}
	handler := loadgavelspace.NewHandler(finder)

	cmd := projects.NewCommand(workspaceOK, handler)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--config", "/tmp/gavel.yaml"})

	err := cmd.Execute()

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	list := out["projects"].([]any)
	require.Len(t, list, 1)
	p := list[0].(map[string]any)
	assert.Equal(t, "core", p["name"])
	assert.Equal(t, "//core/...", p["pattern"])
	gate := p["quality_gate"].(map[string]any)
	rules := gate["rules"].([]any)
	require.Len(t, rules, 1)
	assert.Equal(t, "code_quality", rules[0].(map[string]any)["subtype"])
}

func TestProjectsWorkspaceResolverError(t *testing.T) {
	handler := loadgavelspace.NewHandler(fakeFinder{})
	cmd := projects.NewCommand(workspaceFail, handler)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no workspace")
}

func TestProjectsHandlerError(t *testing.T) {
	finder := fakeFinder{err: errors.New("bad config")}
	handler := loadgavelspace.NewHandler(finder)

	cmd := projects.NewCommand(workspaceOK, handler)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--config", "/tmp/gavel.yaml"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load config")
}

func TestProjectsResolvesConfigPathFromGavelDir(t *testing.T) {
	dir := t.TempDir()
	gavelDir := filepath.Join(dir, ".gavel")
	require.NoError(t, os.MkdirAll(gavelDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(gavelDir, "gavel.yaml"), []byte("name: test"), 0o644))

	gs := newGavelspace(t, "test")
	finder := fakeFinder{gavelspace: gs}
	handler := loadgavelspace.NewHandler(finder)

	resolver := func() (string, error) { return dir, nil }
	cmd := projects.NewCommand(resolver, handler)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	require.NoError(t, cmd.Execute())
}

func TestProjectsResolvesConfigPathFromRoot(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "gavel.yaml"), []byte("name: test"), 0o644))

	gs := newGavelspace(t, "test")
	finder := fakeFinder{gavelspace: gs}
	handler := loadgavelspace.NewHandler(finder)

	resolver := func() (string, error) { return dir, nil }
	cmd := projects.NewCommand(resolver, handler)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	require.NoError(t, cmd.Execute())
}

func TestProjectsFallsBackToGavelDirPath(t *testing.T) {
	dir := t.TempDir()
	finder := fakeFinder{err: errors.New("not found")}
	handler := loadgavelspace.NewHandler(finder)

	resolver := func() (string, error) { return dir, nil }
	cmd := projects.NewCommand(resolver, handler)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load config")
}

func TestProjectsIncludesArchitecturePolicy(t *testing.T) {
	gs := newGavelspace(t, "myrepo")
	proj := newProjectWithArchPolicy(t, "core", "//core/...")
	finder := fakeFinder{gavelspace: gs, projects: []projectmodel.Project{proj}}
	handler := loadgavelspace.NewHandler(finder)

	cmd := projects.NewCommand(workspaceOK, handler)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--config", "/tmp/gavel.yaml"})

	err := cmd.Execute()

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	arch := out["architecture"].(map[string]any)
	layers := arch["layers"].([]any)
	require.Len(t, layers, 2)
	assert.Equal(t, "domain", layers[0].(map[string]any)["name"])
	denyRules := arch["deny_rules"].([]any)
	require.Len(t, denyRules, 1)
	assert.Equal(t, "domain-isolation", denyRules[0].(map[string]any)["name"])
}

func TestRegisterFlagsBindsConfigFlag(t *testing.T) {
	cmd := &cobra.Command{}
	opts := &projects.Options{}

	projects.RegisterFlags(cmd, opts)

	assert.NotNil(t, cmd.Flags().Lookup("config"))
}
