package installer_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/installer"
)

func TestInstall_GoTooling(t *testing.T) {
	dir := t.TempDir()
	inst := installer.NewInstaller()

	_, err := inst.Install(dir, []string{"go"})

	require.NoError(t, err)
	assertGavelBazelrc(t, dir, "go_golangci_lint_submission_aspect")
	assertBazelrcInclude(t, dir)
	assertModuleInclude(t, dir)
	assertBazelDep(t, dir)

	gavelModule := readFile(t, filepath.Join(dir, ".gavel", "gavel.MODULE.bazel"))
	assert.NotContains(t, gavelModule, "golangci_lint_binary",
		"tool binary repos are owned by gavel_tools, not declared by consumers")
}

func TestInstall_JavaTooling(t *testing.T) {
	dir := t.TempDir()
	inst := installer.NewInstaller()

	_, err := inst.Install(dir, []string{"java"})

	require.NoError(t, err)
	gavelBazelrc := readFile(t, filepath.Join(dir, ".gavel", "gavel.bazelrc"))
	assert.Contains(t, gavelBazelrc, "java_pmd_submission_aspect")
	assert.Contains(t, gavelBazelrc, "java_spotbugs_submission_aspect")
	assert.Contains(t, gavelBazelrc, "java_error_prone_submission_aspect")
	assert.Contains(t, gavelBazelrc, "java_cpd_submission_aspect")

	gavelModule := readFile(t, filepath.Join(dir, ".gavel", "gavel.MODULE.bazel"))
	assert.NotContains(t, gavelModule, "_binary")
}

func TestInstall_PythonTooling(t *testing.T) {
	dir := t.TempDir()
	inst := installer.NewInstaller()

	_, err := inst.Install(dir, []string{"python"})

	require.NoError(t, err)
	gavelBazelrc := readFile(t, filepath.Join(dir, ".gavel", "gavel.bazelrc"))
	assert.Contains(t, gavelBazelrc, "python_ruff_submission_aspect")
	assert.Contains(t, gavelBazelrc, "python_bandit_submission_aspect")
	assert.Contains(t, gavelBazelrc, "python_pycompile_submission_aspect")

	gavelModule := readFile(t, filepath.Join(dir, ".gavel", "gavel.MODULE.bazel"))
	assert.NotContains(t, gavelModule, "_binary")
}

func TestInstall_MultiLanguageTooling(t *testing.T) {
	dir := t.TempDir()
	inst := installer.NewInstaller()

	_, err := inst.Install(dir, []string{"go", "java"})

	require.NoError(t, err)
	gavelBazelrc := readFile(t, filepath.Join(dir, ".gavel", "gavel.bazelrc"))
	assert.Contains(t, gavelBazelrc, "go_golangci_lint_submission_aspect")
	assert.Contains(t, gavelBazelrc, "java_pmd_submission_aspect")

	gavelModule := readFile(t, filepath.Join(dir, ".gavel", "gavel.MODULE.bazel"))
	assert.NotContains(t, gavelModule, "_binary")
}

func TestInstall_AppendsTryImportToExistingBazelrc(t *testing.T) {
	dir := t.TempDir()
	existingContent := "build --jobs=8\ntest --test_output=errors\n"
	err := os.WriteFile(filepath.Join(dir, ".bazelrc"), []byte(existingContent), 0o644)
	require.NoError(t, err)

	inst := installer.NewInstaller()

	_, err = inst.Install(dir, []string{"go"})

	require.NoError(t, err)
	bazelrc := readFile(t, filepath.Join(dir, ".bazelrc"))
	assert.Contains(t, bazelrc, "build --jobs=8")
	assert.Contains(t, bazelrc, "try-import %workspace%/.gavel/gavel.bazelrc")
}

func TestInstall_AppendsIncludeToExistingModule(t *testing.T) {
	dir := t.TempDir()
	existingContent := `module(name = "my_project", version = "1.0.0")` + "\n"
	err := os.WriteFile(filepath.Join(dir, "MODULE.bazel"), []byte(existingContent), 0o644)
	require.NoError(t, err)

	inst := installer.NewInstaller()

	_, err = inst.Install(dir, []string{"go"})

	require.NoError(t, err)
	module := readFile(t, filepath.Join(dir, "MODULE.bazel"))
	assert.Contains(t, module, `module(name = "my_project"`)
	assert.Contains(t, module, `include("//:.gavel/gavel.MODULE.bazel")`)
}

func TestInstall_Idempotent(t *testing.T) {
	dir := t.TempDir()
	inst := installer.NewInstaller()

	_, err := inst.Install(dir, []string{"go"})
	require.NoError(t, err)
	firstBazelrc := readFile(t, filepath.Join(dir, ".bazelrc"))
	firstModule := readFile(t, filepath.Join(dir, "MODULE.bazel"))

	_, err = inst.Install(dir, []string{"go"})
	require.NoError(t, err)
	secondBazelrc := readFile(t, filepath.Join(dir, ".bazelrc"))
	secondModule := readFile(t, filepath.Join(dir, "MODULE.bazel"))

	assert.Equal(t, firstBazelrc, secondBazelrc)
	assert.Equal(t, firstModule, secondModule)
}

func TestInstall_ReportsUnchangedOnSecondRun(t *testing.T) {
	dir := t.TempDir()
	inst := installer.NewInstaller()

	first, err := inst.Install(dir, []string{"go"})
	require.NoError(t, err)
	assert.True(t, first[".bazelrc"])
	assert.True(t, first["MODULE.bazel"])

	second, err := inst.Install(dir, []string{"go"})
	require.NoError(t, err)
	assert.False(t, second[".bazelrc"])
	assert.False(t, second["MODULE.bazel"])
}

func TestInstall_DeduplicatesToolingAcrossTargets(t *testing.T) {
	dir := t.TempDir()
	inst := installer.NewInstaller()

	_, err := inst.Install(dir, []string{"go", "go"})

	require.NoError(t, err)
	gavelBazelrc := readFile(t, filepath.Join(dir, ".gavel", "gavel.bazelrc"))
	count := strings.Count(gavelBazelrc, "%go_golangci_lint_submission_aspect")
	assert.Equal(t, 1, count)
}

func TestInstall_SkipsBazelDepForSelfModule(t *testing.T) {
	dir := t.TempDir()
	existingContent := `module(name = "gavel", version = "0.0.0")` + "\n"
	err := os.WriteFile(filepath.Join(dir, "MODULE.bazel"), []byte(existingContent), 0o644)
	require.NoError(t, err)

	inst := installer.NewInstaller()

	_, err = inst.Install(dir, []string{"go"})

	require.NoError(t, err)
	module := readFile(t, filepath.Join(dir, "MODULE.bazel"))
	assert.NotContains(t, module, `bazel_dep(name = "gavel_tools"`)
	assert.Contains(t, module, `include("//:.gavel/gavel.MODULE.bazel")`)
}

func TestVerifyStructure_ValidInstallation(t *testing.T) {
	dir := t.TempDir()
	inst := installer.NewInstaller()

	_, err := inst.Install(dir, []string{"go"})
	require.NoError(t, err)

	issues, err := inst.VerifyStructure(dir)

	require.NoError(t, err)
	assert.Empty(t, issues)
}

func TestVerifyStructure_MissingFiles(t *testing.T) {
	dir := t.TempDir()

	inst := installer.NewInstaller()

	issues, err := inst.VerifyStructure(dir)

	require.NoError(t, err)
	require.NotEmpty(t, issues)
	assert.Contains(t, issues[0], "gavel.bazelrc not found")
	assert.Contains(t, issues[1], "gavel.MODULE.bazel not found")
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

func assertGavelBazelrc(t *testing.T, dir, expectedContent string) {
	t.Helper()
	content := readFile(t, filepath.Join(dir, ".gavel", "gavel.bazelrc"))
	assert.Contains(t, content, expectedContent)
}

func assertBazelrcInclude(t *testing.T, dir string) {
	t.Helper()
	content := readFile(t, filepath.Join(dir, ".bazelrc"))
	assert.Contains(t, content, "try-import %workspace%/.gavel/gavel.bazelrc")
}

func assertModuleInclude(t *testing.T, dir string) {
	t.Helper()
	content := readFile(t, filepath.Join(dir, "MODULE.bazel"))
	assert.Contains(t, content, `include("//:.gavel/gavel.MODULE.bazel")`)
}

func assertBazelDep(t *testing.T, dir string) {
	t.Helper()
	content := readFile(t, filepath.Join(dir, "MODULE.bazel"))
	assert.Contains(t, content, `bazel_dep(name = "gavel_tools"`)
}
