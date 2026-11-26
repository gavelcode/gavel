package installer_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/gavelspace/loadgavelspace"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/installer"
)

const gavelDepLine = `bazel_dep(name = "gavel_tools", version = "0.4.0")`

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

func TestInstall_BazelrcUsesEachToolsModule(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"go", "java", "python", "typescript", "rust"})
	require.NoError(t, err)

	bazelrc := readFile(t, filepath.Join(root, installer.GavelBazelrc))
	for _, want := range []string{
		"@gavel_tools_go//lint/aspects:defs.bzl%go_golangci_lint_submission_aspect",
		"@gavel_tools_go//lint/aspects:defs.bzl%go_archtest_submission_aspect",
		"@gavel_tools_java//lint/aspects:defs.bzl%java_pmd_submission_aspect",
		"@gavel_tools_python//lint/aspects:defs.bzl%python_ruff_submission_aspect",
		"@gavel_tools_typescript//lint/aspects:defs.bzl%typescript_eslint_submission_aspect",
		"@gavel_tools_rust//lint/aspects:defs.bzl%rust_clippy_submission_aspect",
		"@gavel_tools//lint/aspects:defs.bzl%rust_archtest_submission_aspect",
	} {
		assert.Contains(t, bazelrc, want)
	}
	assert.NotContains(t, bazelrc, "@gavel//")
}

func moduleDepNames(t *testing.T, root string) []string {
	t.Helper()
	var names []string
	for _, match := range regexp.MustCompile(`bazel_dep\(name = "([^"]+)"`).FindAllStringSubmatch(readRaw(t, filepath.Join(root, installer.GavelModule)), -1) {
		names = append(names, match[1])
	}
	return names
}

func TestInstall_MultiLanguage_WritesOneDepPerDistinctModuleSorted(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"typescript", "go", "java", "python", "rust"})
	require.NoError(t, err)

	assert.Equal(t, []string{"gavel_tools_go", "gavel_tools_java", "gavel_tools_python", "gavel_tools_rust", "gavel_tools_typescript"},
		moduleDepNames(t, root), "core gavel_tools stays in the root MODULE.bazel, not duplicated here")
	assert.Contains(t, readFile(t, filepath.Join(root, "MODULE.bazel")), gavelDepLine)
	assert.Contains(t, readFile(t, filepath.Join(root, installer.GavelModule)), `bazel_dep(name = "gavel_tools_go", version = "0.4.0")`)
}

func TestInstall_JavaOnly_WritesOnlyJavaModule(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"java"})
	require.NoError(t, err)

	assert.Equal(t, []string{"gavel_tools_java"}, moduleDepNames(t, root))
	bazelrc := readFile(t, filepath.Join(root, installer.GavelBazelrc))
	assert.NotContains(t, bazelrc, "gavel_tools_go")
	assert.NotContains(t, bazelrc, "gavel_tools_rust")
}

func TestInstall_RustWithoutRulesRs_UsesBaseRustModule(t *testing.T) {
	root := setupWorkspace(t, `bazel_dep(name = "rules_rust", version = "0.60.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"rust"})
	require.NoError(t, err)

	assert.Equal(t, []string{"gavel_tools_rust"}, moduleDepNames(t, root))
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

func TestInstall_NoVariantDeps_OutputMatchesGolden(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n"+
		`bazel_dep(name = "rules_rust", version = "0.60.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"go", "java", "python", "typescript", "rust"})
	require.NoError(t, err)

	for file, golden := range map[string]string{
		installer.GavelBazelrc: "testdata/gavel.bazelrc.golden",
		installer.GavelModule:  "testdata/gavel.MODULE.bazel.golden",
	} {
		want, err := os.ReadFile(golden)
		require.NoError(t, err)
		assert.Equal(t, string(want), readRaw(t, filepath.Join(root, file)),
			"consumers without a matching ruleset must get exactly the base wiring")
	}
}

func TestInstall_RulesRsConsumer_GetsVariantDepAndAspect(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n"+
		`bazel_dep(name = "rules_rs", version = "0.2.4")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"rust"})
	require.NoError(t, err)

	module := readFile(t, filepath.Join(root, installer.GavelModule))
	assert.Contains(t, module, `bazel_dep(name = "gavel_tools_rust_rs", version = "0.4.0")`)
	assert.Equal(t, []string{"gavel_tools_rust_rs"}, moduleDepNames(t, root),
		"the variant replaces gavel_tools_rust; archtest's core gavel_tools lives in the root MODULE.bazel")

	bazelrc := readFile(t, filepath.Join(root, installer.GavelBazelrc))
	assert.Contains(t, bazelrc, "build:gavel-rust-clippy --aspects=@gavel_tools_rust_rs//lint/aspects:defs.bzl%rust_clippy_submission_aspect")
	assert.Contains(t, bazelrc, "build:gavel-rust-archtest --aspects=@gavel_tools//lint/aspects:defs.bzl%rust_archtest_submission_aspect")
}

func TestInstall_RulesRsConsumer_ForRustOnly(t *testing.T) {
	root := setupWorkspace(t, `bazel_dep(name = "rules_rs", version = "0.2.4")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	assert.NotContains(t, readFile(t, filepath.Join(root, installer.GavelModule)), "gavel_tools_rust_rs",
		"a variant module is only added when its language is selected")
}

func TestInstall_RulesRsConsumer_IsIdempotent(t *testing.T) {
	root := setupWorkspace(t, `bazel_dep(name = "rules_rs", version = "0.2.4")`+"\n")
	inst := installer.NewInstaller()

	_, err := inst.Install(root, []string{"rust"})
	require.NoError(t, err)
	first := readRaw(t, filepath.Join(root, installer.GavelModule)) + readRaw(t, filepath.Join(root, installer.GavelBazelrc)) +
		readRaw(t, filepath.Join(root, "MODULE.bazel"))

	_, err = inst.Install(root, []string{"rust"})
	require.NoError(t, err)
	second := readRaw(t, filepath.Join(root, installer.GavelModule)) + readRaw(t, filepath.Join(root, installer.GavelBazelrc)) +
		readRaw(t, filepath.Join(root, "MODULE.bazel"))

	assert.Equal(t, first, second)
	assert.Equal(t, 1, strings.Count(second, `name = "gavel_tools_rust_rs"`))
}

func readRaw(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

func shellcheckBridge() loadgavelspace.BridgeView {
	return loadgavelspace.BridgeView{
		Name: "shellcheck", Aspect: "//tools/lint:shellcheck.bzl%shellcheck_gavel", SARIFSuffix: ".shellcheck.sarif",
		BuildFlags: []string{"--some=flag", "--other"},
	}
}

func TestInstall_Bridges_WritesOneBazelrcConfigPerBridge(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"go"}, shellcheckBridge())
	require.NoError(t, err)

	bazelrc := readFile(t, filepath.Join(root, installer.GavelBazelrc))
	for _, want := range []string{
		"# --- Bridges (provided by the project, not guaranteed by gavel) ---",
		"build:gavel-bridge-shellcheck --aspects=//tools/lint:shellcheck.bzl%shellcheck_gavel",
		"build:gavel-bridge-shellcheck --output_groups=gavel_submissions",
		"build:gavel-bridge-shellcheck --some=flag",
		"build:gavel-bridge-shellcheck --other",
	} {
		assert.Contains(t, bazelrc, want)
	}
}

func TestInstall_Bridges_AddNoBazelDep(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")
	plain := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, nil, shellcheckBridge())
	require.NoError(t, err)
	_, err = installer.NewInstaller().Install(plain, nil)
	require.NoError(t, err)

	assert.Equal(t, readRaw(t, filepath.Join(plain, installer.GavelModule)), readRaw(t, filepath.Join(root, installer.GavelModule)))
	assert.Equal(t, readRaw(t, filepath.Join(plain, "MODULE.bazel")), readRaw(t, filepath.Join(root, "MODULE.bazel")))
}

func TestInstall_Bridges_IsIdempotent(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")
	inst := installer.NewInstaller()

	_, err := inst.Install(root, []string{"go"}, shellcheckBridge())
	require.NoError(t, err)
	first := readRaw(t, filepath.Join(root, installer.GavelBazelrc)) + readRaw(t, filepath.Join(root, ".bazelrc"))
	_, err = inst.Install(root, []string{"go"}, shellcheckBridge())
	require.NoError(t, err)
	second := readRaw(t, filepath.Join(root, installer.GavelBazelrc)) + readRaw(t, filepath.Join(root, ".bazelrc"))

	assert.Equal(t, first, second)
}

func TestInstall_Bridges_SameBridgeDeclaredByTwoProjectsIsWrittenOnce(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, nil, shellcheckBridge(), shellcheckBridge())
	require.NoError(t, err)

	assert.Equal(t, 1, strings.Count(readRaw(t, filepath.Join(root, installer.GavelBazelrc)), "build:gavel-bridge-shellcheck --aspects="))
}

func TestInstall_Bridges_RejectsSameNameWithDifferentDefinitions(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")
	other := shellcheckBridge()
	other.Aspect = "//tools/lint:other.bzl%other"

	_, err := installer.NewInstaller().Install(root, nil, shellcheckBridge(), other)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "shellcheck")
}

func TestInstall_NoBridges_OmitsBridgeSection(t *testing.T) {
	root := setupWorkspace(t, `module(name = "consumer-app", version = "1.0.0")`+"\n")

	_, err := installer.NewInstaller().Install(root, []string{"go"})
	require.NoError(t, err)

	assert.NotContains(t, readRaw(t, filepath.Join(root, installer.GavelBazelrc)), "Bridges")
}
