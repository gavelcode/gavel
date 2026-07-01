package catalog_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/catalog"
)

const testCatalogYAML = `
version: 1
aspects_bzl: //lint/aspects:defs.bzl
languages:
  go:
    - {name: golangci-lint, aspect: go_golangci_lint_submission_aspect, sarif_suffix: .golangci.sarif, build_flags: ["--@rules_go//go/config:export_stdlib=True"], binary: golangci_lint_binary}
    - {name: archtest, aspect: go_archtest_submission_aspect, sarif_suffix: .archtest.sarif}
  java:
    - {name: pmd, aspect: java_pmd_submission_aspect, sarif_suffix: .pmd.sarif, binary: pmd_binary}
    - {name: cpd, aspect: java_cpd_submission_aspect, sarif_suffix: .cpd.sarif, binary: pmd_binary}
    - {name: spotbugs, aspect: java_spotbugs_submission_aspect, sarif_suffix: .spotbugs.sarif, binary: spotbugs_binary}
    - {name: error-prone, aspect: java_error_prone_submission_aspect, sarif_suffix: .errorprone.sarif, binary: error_prone_binary}
    - {name: archtest, aspect: java_archtest_submission_aspect, sarif_suffix: .archtest.sarif}
  python:
    - {name: ruff, aspect: python_ruff_submission_aspect, sarif_suffix: .ruff.sarif, binary: ruff_binary}
    - {name: bandit, aspect: python_bandit_submission_aspect, sarif_suffix: .bandit.sarif, binary: bandit_binary}
    - {name: pycompile, aspect: python_pycompile_submission_aspect, sarif_suffix: .pycompile.sarif}
    - {name: archtest, aspect: python_archtest_submission_aspect, sarif_suffix: .archtest.sarif}
  typescript:
    - {name: eslint, aspect: typescript_eslint_submission_aspect, sarif_suffix: .eslint.sarif}
    - {name: archtest, aspect: typescript_archtest_submission_aspect, sarif_suffix: .archtest.sarif}
  rust:
    - {name: clippy, aspect: rust_clippy_submission_aspect, sarif_suffix: .clippy.sarif}
    - {name: archtest, aspect: rust_archtest_submission_aspect, sarif_suffix: .archtest.sarif}
`

func TestMain(m *testing.M) {
	loadTestCatalog()
	os.Exit(m.Run())
}

func loadTestCatalog() {
	parsed, err := catalog.ParseCatalog([]byte(testCatalogYAML))
	if err != nil {
		panic(err)
	}
	catalog.SetCatalog(parsed)
}

func TestParseCatalog_RejectsMalformedYAML(t *testing.T) {
	_, err := catalog.ParseCatalog([]byte("languages: [::"))
	assert.Error(t, err)
}

func TestParseCatalog_RejectsMissingAspectsBzl(t *testing.T) {
	_, err := catalog.ParseCatalog([]byte("version: 1\nlanguages:\n  go:\n    - {name: x, aspect: y}\n"))
	assert.ErrorContains(t, err, "aspects_bzl")
}

func TestParseCatalog_RejectsNoLanguages(t *testing.T) {
	_, err := catalog.ParseCatalog([]byte("version: 1\naspects_bzl: //a:b\n"))
	assert.ErrorContains(t, err, "no languages")
}

func TestParseCatalog_RejectsToolWithoutAspect(t *testing.T) {
	_, err := catalog.ParseCatalog([]byte("version: 1\naspects_bzl: //a:b\nlanguages:\n  go:\n    - {name: x}\n"))
	assert.ErrorContains(t, err, "no aspect")
}

func TestRuleKindsForLanguage_Go(t *testing.T) {
	kinds := catalog.RuleKindsForLanguage("go")
	assert.Contains(t, kinds, "go_library")
	assert.Contains(t, kinds, "go_binary")
	assert.Contains(t, kinds, "go_test")
}

func TestRuleKindsForLanguage_Unknown(t *testing.T) {
	assert.Empty(t, catalog.RuleKindsForLanguage("cobol"))
}

func TestLintAspectsForLanguages_Go(t *testing.T) {
	aspects := catalog.LintAspectsForLanguages([]string{"go"})

	assert.Len(t, aspects, 1)
	assert.Equal(t, "go_golangci_lint_submission_aspect", aspects[0].Name)
	assert.Equal(t, "@gavel//lint/aspects:defs.bzl%go_golangci_lint_submission_aspect", aspects[0].Path)
}

func TestLintAspectsForLanguages_GoGolangciCarriesExportStdlibFlag(t *testing.T) {
	aspects := catalog.LintAspectsForLanguages([]string{"go"})

	assert.Equal(t, []string{"--@rules_go//go/config:export_stdlib=True"}, aspects[0].BuildFlags,
		"golangci runs hermetic via the packages driver, which needs stdlib export data")
}

func TestLintAspectsForLanguages_NonGoAspectsHaveNoBuildFlags(t *testing.T) {
	for _, aspect := range catalog.LintAspectsForLanguages([]string{"java", "python", "rust", "typescript"}) {
		assert.Empty(t, aspect.BuildFlags, "%s must not force stdlib export — it is go-only and costly", aspect.Name)
	}
}

func TestLintAspectsForLanguages_DeduplicatesSharedAspects(t *testing.T) {
	shared := `
version: 1
aspects_bzl: //lint/aspects:defs.bzl
languages:
  java:   [{name: pmd, aspect: java_pmd_submission_aspect, sarif_suffix: .pmd.sarif}]
  kotlin: [{name: pmd, aspect: java_pmd_submission_aspect, sarif_suffix: .pmd.sarif}]
`
	parsed, err := catalog.ParseCatalog([]byte(shared))
	require.NoError(t, err)
	catalog.SetCatalog(parsed)
	t.Cleanup(loadTestCatalog)

	aspects := catalog.LintAspectsForLanguages([]string{"java", "kotlin"})
	assert.Len(t, aspects, 1, "an aspect shared across languages is selected once")
}

func TestLintAspectsForLanguages_EmptyReturnsNothing(t *testing.T) {
	assert.Empty(t, catalog.LintAspectsForLanguages(nil))
	assert.Empty(t, catalog.LintAspectsForLanguages([]string{}))
}

func TestLintAspectsForLanguages_UnknownLanguage(t *testing.T) {
	assert.Empty(t, catalog.LintAspectsForLanguages([]string{"cobol"}))
}

func TestSelectedAspects_ResolvesChosenToolsPerLanguage(t *testing.T) {
	aspects, err := catalog.SelectedAspects(map[string][]string{"go": {"golangci-lint", "archtest"}})

	require.NoError(t, err)
	require.Len(t, aspects, 2)
	assert.Equal(t, "go_golangci_lint_submission_aspect", aspects[0].Name)
	assert.Equal(t, "go_archtest_submission_aspect", aspects[1].Name)
	assert.Equal(t, []string{"--@rules_go//go/config:export_stdlib=True"}, aspects[0].BuildFlags)
}

func TestSelectedAspects_PreservesSelectionOrder(t *testing.T) {
	aspects, err := catalog.SelectedAspects(map[string][]string{"go": {"archtest", "golangci-lint"}})

	require.NoError(t, err)
	require.Len(t, aspects, 2)
	assert.Equal(t, "go_archtest_submission_aspect", aspects[0].Name, "the chosen order is preserved")
}

func TestSelectedAspects_OrdersLanguagesDeterministically(t *testing.T) {
	aspects, err := catalog.SelectedAspects(map[string][]string{"python": {"ruff"}, "go": {"golangci-lint"}})

	require.NoError(t, err)
	require.Len(t, aspects, 2)
	assert.Equal(t, "go_golangci_lint_submission_aspect", aspects[0].Name, "languages sorted alphabetically")
	assert.Equal(t, "python_ruff_submission_aspect", aspects[1].Name)
}

func TestSelectedAspects_ErrorsOnUnknownTool(t *testing.T) {
	_, err := catalog.SelectedAspects(map[string][]string{"go": {"clang-tidy"}})

	assert.ErrorContains(t, err, "clang-tidy")
	assert.ErrorContains(t, err, "go")
}

func TestSelectedAspects_ErrorsOnUnknownLanguage(t *testing.T) {
	_, err := catalog.SelectedAspects(map[string][]string{"cobol": {"anything"}})

	assert.ErrorContains(t, err, "anything")
}

func TestSelectedAspects_EmptySelectionReturnsNothing(t *testing.T) {
	aspects, err := catalog.SelectedAspects(map[string][]string{})

	require.NoError(t, err)
	assert.Empty(t, aspects)
}

func TestAspectNames_AllLanguages(t *testing.T) {
	names := catalog.AspectNames([]string{"go", "java", "python", "typescript", "rust"})

	assert.Contains(t, names, "go_golangci_lint_submission_aspect")
	assert.Contains(t, names, "java_pmd_submission_aspect")
	assert.Contains(t, names, "python_ruff_submission_aspect")
	assert.Contains(t, names, "typescript_eslint_submission_aspect")
	assert.Contains(t, names, "rust_clippy_submission_aspect")
}

func TestAspectNames_EmptyReturnsNothing(t *testing.T) {
	assert.Empty(t, catalog.AspectNames(nil))
	assert.Empty(t, catalog.AspectNames([]string{}))
}

func TestAspectNames_SingleLanguage(t *testing.T) {
	names := catalog.AspectNames([]string{"go"})

	assert.Equal(t, []string{"go_golangci_lint_submission_aspect", "go_archtest_submission_aspect"}, names)
}

func TestBinaryNames_AllLanguages(t *testing.T) {
	names := catalog.BinaryNames([]string{"go", "java", "python", "typescript", "rust"})

	assert.Equal(t, []string{
		"golangci_lint_binary",
		"pmd_binary",
		"spotbugs_binary",
		"error_prone_binary",
		"ruff_binary",
		"bandit_binary",
	}, names, "cpd shares pmd_binary, so it is de-duplicated")
}

func TestBinaryNames_EmptyReturnsNothing(t *testing.T) {
	assert.Empty(t, catalog.BinaryNames(nil))
	assert.Empty(t, catalog.BinaryNames([]string{}))
}

func TestBinaryNames_SingleLanguage(t *testing.T) {
	assert.Equal(t, []string{"golangci_lint_binary"}, catalog.BinaryNames([]string{"go"}))
}

func TestBinaryNames_LanguageWhoseToolsDeclareNoBinary(t *testing.T) {
	assert.Empty(t, catalog.BinaryNames([]string{"typescript"}))
}

func TestArchtestAspectsForLanguages_OnlyArchtest(t *testing.T) {
	aspects := catalog.ArchtestAspectsForLanguages([]string{"go"})

	assert.Len(t, aspects, 1)
	assert.Equal(t, "go_archtest_submission_aspect", aspects[0].Name)
}

func TestIsArchtestAspect(t *testing.T) {
	assert.True(t, catalog.IsArchtestAspect("go_archtest_submission_aspect"))
	assert.False(t, catalog.IsArchtestAspect("go_golangci_lint_submission_aspect"))
}

func TestAllAspectsForLanguages_Go(t *testing.T) {
	aspects := catalog.AllAspectsForLanguages([]string{"go"})

	assert.Len(t, aspects, 2)
	names := make([]string, len(aspects))
	for i, aspect := range aspects {
		names[i] = aspect.Name
	}
	assert.Contains(t, names, "go_golangci_lint_submission_aspect")
	assert.Contains(t, names, "go_archtest_submission_aspect")
}

func TestAllAspectsForLanguages_Java(t *testing.T) {
	assert.Len(t, catalog.AllAspectsForLanguages([]string{"java"}), 5)
}

func TestAspectPaths_FormatsCommaSeparated(t *testing.T) {
	aspects := catalog.LintAspectsForLanguages([]string{"go"})

	assert.Equal(t, "@gavel//lint/aspects:defs.bzl%go_golangci_lint_submission_aspect", catalog.AspectPaths(aspects))
}

func TestAspectPaths_MultipleAspects(t *testing.T) {
	got := catalog.AspectPaths(catalog.AllAspectsForLanguages([]string{"go"}))

	assert.Contains(t, got, ",")
	assert.Contains(t, got, "go_golangci_lint_submission_aspect")
	assert.Contains(t, got, "go_archtest_submission_aspect")
}

func TestAspectPaths_EmptySlice(t *testing.T) {
	assert.Equal(t, "", catalog.AspectPaths(nil))
}

func TestModulePrefix_DefaultIsGavel(t *testing.T) {
	assert.Equal(t, "@gavel", catalog.ModulePrefix())
}

func TestSetModulePrefix_ChangesAspectPaths(t *testing.T) {
	catalog.SetModulePrefix("@gavel_tools")
	t.Cleanup(func() { catalog.SetModulePrefix("@gavel") })

	aspects := catalog.AllAspectsForLanguages([]string{"go"})

	assert.Equal(t, "@gavel_tools//lint/aspects:defs.bzl%go_golangci_lint_submission_aspect", aspects[0].Path)
	assert.Equal(t, "@gavel_tools//lint/aspects:defs.bzl%go_archtest_submission_aspect", aspects[1].Path)
}
