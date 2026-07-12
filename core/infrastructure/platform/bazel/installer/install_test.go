package installer_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/catalog"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/installer"
)

const gavelDepLine = `bazel_dep(name = "gavel_tools", version = "0.3.8")`

const (
	gavelRegistryLine = "common --registry=https://gavelcode.github.io/registry"
	bcrRegistryLine   = "common --registry=https://bcr.bazel.build"
)

func TestInstall_DownstreamRepo_InjectsBazelDep(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	module := readFile(t, filepath.Join(root, "MODULE.bazel"))
	assert.Contains(t, module, gavelDepLine,
		"downstream repos must depend on gavel via bazel_dep")
}

func TestInstall_GavelProjectSingleLine_DoesNotInjectBazelDep(t *testing.T) {
	root := setupWorkspace(t, `module(name = "gavel", version = "0.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	module := readFile(t, filepath.Join(root, "MODULE.bazel"))
	assert.NotContains(t, module, gavelDepLine,
		"gavel-project itself must not declare bazel_dep on its own module name")
}

func TestInstall_GavelProjectMultiLine_DoesNotInjectBazelDep(t *testing.T) {
	root := setupWorkspace(t, "module(\n    name = \"gavel\",\n    version = \"0.0.0\",\n)\n")

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	module := readFile(t, filepath.Join(root, "MODULE.bazel"))
	assert.NotContains(t, module, gavelDepLine,
		"multi-line module() form must also be detected as self-module")
}

func TestInstall_GavelToolsProject_DoesNotInjectBazelDep(t *testing.T) {
	root := setupWorkspace(t, "module(\n    name = \"gavel_tools\",\n    version = \"0.1.0\",\n)\n")

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	module := readFile(t, filepath.Join(root, "MODULE.bazel"))
	assert.NotContains(t, module, gavelDepLine,
		"gavel_tools itself must not declare a bazel_dep on its own module name")
}

func TestInstall_AnyShape_AlwaysAddsInclude(t *testing.T) {
	cases := []struct {
		name    string
		initial string
	}{
		{"downstream", `module(name = "consumer-app", version = "1.0.0")` + "\n"},
		{"self-single-line", `module(name = "gavel", version = "0.0.0")` + "\n"},
		{"self-multi-line", "module(\n    name = \"gavel\",\n    version = \"0.0.0\",\n)\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := setupWorkspace(t, tc.initial)

			_, err := installer.NewInstaller().Install(root, []string{"go"})
			require.NoError(t, err)

			module := readFile(t, filepath.Join(root, "MODULE.bazel"))
			assert.Contains(t, module, `include("//:.gavel/gavel.MODULE.bazel")`,
				"every workspace must include the generated gavel MODULE.bazel")
		})
	}
}

func TestInstall_AddsLintConfigFilegroup(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	build := readFile(t, filepath.Join(root, "BUILD.bazel"))
	assert.Contains(t, build, "gavel_lint_config")
	assert.Contains(t, build, ".golangci.yml")
	assert.Contains(t, build, ".gavel/architecture.yml")
}

func TestInstall_DoesNotDuplicateFilegroup(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	_, err = installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	build := readFile(t, filepath.Join(root, "BUILD.bazel"))
	assert.Equal(t, 1, strings.Count(build, "gavel_lint_config"),
		"filegroup must not be duplicated on repeated installs")
}

func TestInstall_ExistingRootBuild_AppendsWithoutShadowing(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")
	original := "config_setting(\n    name = \"fastbuild\",\n    values = {\"compilation_mode\": \"fastbuild\"},\n)\n"
	require.NoError(t, os.WriteFile(filepath.Join(root, "BUILD"), []byte(original), 0o644))

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	_, statErr := os.Stat(filepath.Join(root, "BUILD.bazel"))
	assert.True(t, os.IsNotExist(statErr), "must not create a BUILD.bazel that shadows the existing root BUILD")

	build := readFile(t, filepath.Join(root, "BUILD"))
	assert.Contains(t, build, "gavel_lint_config", "filegroup must land in the existing root BUILD")
	assert.Contains(t, build, "fastbuild", "the existing root package must be preserved")
}

func TestInstall_BothRootBuildFiles_PrefersBuildBazel(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")
	require.NoError(t, os.WriteFile(filepath.Join(root, "BUILD"), []byte("# legacy\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "BUILD.bazel"), []byte("# active\n"), 0o644))

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	assert.Contains(t, readFile(t, filepath.Join(root, "BUILD.bazel")), "gavel_lint_config",
		"when both exist, Bazel uses BUILD.bazel, so the filegroup belongs there")
	assert.NotContains(t, readFile(t, filepath.Join(root, "BUILD")), "gavel_lint_config")
}

func TestInstall_AddsGavelRegistry(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	bazelrc := readFile(t, filepath.Join(root, ".bazelrc"))
	assert.Contains(t, bazelrc, gavelRegistryLine,
		"init must point .bazelrc at the gavel registry so gavel_tools resolves")
}

func TestInstall_KeepsBcrAlongsideGavelRegistry(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	bazelrc := readFile(t, filepath.Join(root, ".bazelrc"))
	assert.Contains(t, bazelrc, bcrRegistryLine,
		"declaring any registry drops Bazel's BCR default, so BCR must be added explicitly")
}

func TestInstall_BcrRegistryOrderedBeforeGavel(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	bazelrc := readFile(t, filepath.Join(root, ".bazelrc"))
	bcrAt := strings.Index(bazelrc, bcrRegistryLine)
	gavelAt := strings.Index(bazelrc, gavelRegistryLine)
	require.NotEqual(t, -1, bcrAt)
	require.NotEqual(t, -1, gavelAt)
	assert.Less(t, bcrAt, gavelAt,
		"BCR must precede the gavel registry so the consumer's own BCR modules keep resolving from BCR with unchanged checksums; listing gavel first re-resolves every module through it and breaks repos with lockfile_mode=error")
}

func TestInstall_ExistingBcr_NotDuplicated(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")
	require.NoError(t, os.WriteFile(filepath.Join(root, ".bazelrc"), []byte(bcrRegistryLine+"\n"), 0o644))

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	bazelrc := readFile(t, filepath.Join(root, ".bazelrc"))
	assert.Contains(t, bazelrc, gavelRegistryLine)
	assert.Equal(t, 1, strings.Count(bazelrc, bcrRegistryLine),
		"must not duplicate an existing BCR registry line")
}

func TestInstall_DoesNotDuplicateGavelRegistry(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)
	_, err = installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	bazelrc := readFile(t, filepath.Join(root, ".bazelrc"))
	assert.Equal(t, 1, strings.Count(bazelrc, gavelRegistryLine),
		"registry line must not be duplicated on repeated installs")
}

func TestVerifyStructure_AllPresent(t *testing.T) {
	root := setupWorkspace(t, `module(name = "app", version = "1.0.0")`+"\n")
	inst := installer.NewInstaller()
	_, err := inst.Install(root, []string{"go"})
	require.NoError(t, err)

	issues, err := inst.VerifyStructure(root)
	require.NoError(t, err)
	assert.Empty(t, issues)
}

func TestVerifyStructure_MissingGavelBazelrc(t *testing.T) {
	root := setupWorkspace(t, `module(name = "app", version = "1.0.0")`+"\n")
	inst := installer.NewInstaller()
	_, err := inst.Install(root, []string{"go"})
	require.NoError(t, err)

	require.NoError(t, os.Remove(filepath.Join(root, installer.GavelBazelrc)))

	issues, err := inst.VerifyStructure(root)
	require.NoError(t, err)
	assert.Contains(t, strings.Join(issues, "\n"), installer.GavelBazelrc+" not found")
}

func TestVerifyStructure_MissingBazelrc(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".gavel"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, installer.GavelBazelrc), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, installer.GavelModule), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "MODULE.bazel"), []byte(`include("//:.gavel/gavel.MODULE.bazel")`), 0o644))

	inst := installer.NewInstaller()
	issues, err := inst.VerifyStructure(root)
	require.NoError(t, err)
	assert.Contains(t, strings.Join(issues, "\n"), ".bazelrc not found")
}

func TestVerifyStructure_MissingIncludeLine(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".gavel"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, installer.GavelBazelrc), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, installer.GavelModule), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".bazelrc"), []byte("# empty\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "MODULE.bazel"), []byte(`module(name = "app")`+"\n"), 0o644))

	inst := installer.NewInstaller()
	issues, err := inst.VerifyStructure(root)
	require.NoError(t, err)
	assert.Len(t, issues, 2, "both .bazelrc and MODULE.bazel should report missing include lines")
}

func TestCatalog_ReturnsAspectAndBinaryNames(t *testing.T) {
	inst := installer.NewInstaller()

	aspects, binaries := inst.Catalog([]string{"go"})

	assert.NotEmpty(t, aspects)
	assert.NotEmpty(t, binaries)
}

func TestInstall_MultipleLanguages(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	modified, err := installer.NewInstaller().Install(root, []string{"go", "java", "python", "typescript", "rust"})

	require.NoError(t, err)
	assert.True(t, modified[installer.GavelBazelrc])
	assert.True(t, modified[installer.GavelModule])
}

func TestInstall_ExistingBuildFile_AppendsFilegroup(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")
	require.NoError(t, os.WriteFile(filepath.Join(root, "BUILD.bazel"), []byte("load(\"//rules:go.bzl\", \"go_library\")\n"), 0o644))

	modified, err := installer.NewInstaller().Install(root, []string{"go"})

	require.NoError(t, err)
	assert.True(t, modified["BUILD.bazel"])
	build := readFile(t, filepath.Join(root, "BUILD.bazel"))
	assert.Contains(t, build, "go_library")
	assert.Contains(t, build, "gavel_lint_config")
}

func TestInstall_ErrorWhenGavelDirBlocked(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "MODULE.bazel"), []byte(`module(name = "app")`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".gavel"), []byte("block"), 0o644))

	_, err := installer.NewInstaller().Install(root, []string{"go"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "create .gavel directory")
}

func TestInstall_ErrorWhenBazelrcWriteFails(t *testing.T) {
	root := setupWorkspace(t, `module(name = "app")`+"\n")
	gavelDir := filepath.Join(root, ".gavel")
	require.NoError(t, os.MkdirAll(gavelDir, 0o755))
	require.NoError(t, os.Chmod(gavelDir, 0o555))
	t.Cleanup(func() { _ = os.Chmod(gavelDir, 0o755) })

	_, err := installer.NewInstaller().Install(root, []string{"go"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "install bazelrc")
}

func TestInstall_ErrorWhenBazelrcUnreadable(t *testing.T) {
	root := setupWorkspace(t, `module(name = "app")`+"\n")
	require.NoError(t, os.Chmod(filepath.Join(root, ".bazelrc"), 0o000))
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(root, ".bazelrc"), 0o644) })

	_, err := installer.NewInstaller().Install(root, []string{"go"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "install bazelrc")
}

func TestInstall_ErrorWhenModuleUnreadable(t *testing.T) {
	root := setupWorkspace(t, `module(name = "app")`+"\n")
	require.NoError(t, os.Chmod(filepath.Join(root, "MODULE.bazel"), 0o000))
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(root, "MODULE.bazel"), 0o644) })

	_, err := installer.NewInstaller().Install(root, []string{"go"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "install module")
}

func TestInstall_ErrorWhenBuildFileUnreadable(t *testing.T) {
	root := setupWorkspace(t, `module(name = "app")`+"\n")
	require.NoError(t, os.WriteFile(filepath.Join(root, "BUILD.bazel"), []byte("existing"), 0o000))
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(root, "BUILD.bazel"), 0o644) })

	_, err := installer.NewInstaller().Install(root, []string{"go"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "install lint config filegroup")
}

func TestVerifyStructure_MissingGavelModule(t *testing.T) {
	root := setupWorkspace(t, `module(name = "app", version = "1.0.0")`+"\n")
	inst := installer.NewInstaller()
	_, err := inst.Install(root, []string{"go"})
	require.NoError(t, err)

	require.NoError(t, os.Remove(filepath.Join(root, installer.GavelModule)))

	issues, err := inst.VerifyStructure(root)

	require.NoError(t, err)
	assert.Contains(t, strings.Join(issues, "\n"), installer.GavelModule+" not found")
}

func TestInstall_BazelrcUsesModulePrefix(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	bazelrc := readFile(t, filepath.Join(root, installer.GavelBazelrc))
	assert.Contains(t, bazelrc, "@gavel//lint/aspects:defs.bzl%go_golangci_lint_submission_aspect")
	assert.Contains(t, bazelrc, "@gavel//lint/aspects:defs.bzl%go_archtest_submission_aspect")
}

func TestInstall_BazelrcReflectsCustomModulePrefix(t *testing.T) {
	catalog.SetModulePrefix("@gavel_tools")
	t.Cleanup(func() { catalog.SetModulePrefix("@gavel") })

	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	bazelrc := readFile(t, filepath.Join(root, installer.GavelBazelrc))
	assert.Contains(t, bazelrc, "@gavel_tools//lint/aspects:defs.bzl%go_golangci_lint_submission_aspect")
	assert.Contains(t, bazelrc, "@gavel_tools//lint/aspects:defs.bzl%go_archtest_submission_aspect")
	assert.NotContains(t, bazelrc, "@gavel//lint/aspects")
}

func TestInstall_ModuleBazelDoesNotDeclareToolRepos(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"go", "java", "python"})
	require.NoError(t, err)

	module := readFile(t, filepath.Join(root, installer.GavelModule))
	assert.NotContains(t, module, "repositories.bzl")
	assert.NotContains(t, module, "use_repo_rule")
}

func setupWorkspace(t *testing.T, moduleContent string) string {
	t.Helper()
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "MODULE.bazel"), []byte(moduleContent), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".bazelrc"), []byte(""), 0o644))
	return root
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return strings.TrimSpace(string(data))
}
