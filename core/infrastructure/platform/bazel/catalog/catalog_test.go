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
    - {module: gavel_tools_go, name: golangci-lint, aspect: go_golangci_lint_submission_aspect, sarif_suffix: .golangci.sarif, build_flags: ["--@rules_go//go/config:export_stdlib=True"], binary: golangci_lint_binary}
    - {module: gavel_tools_go, name: archtest, aspect: go_archtest_submission_aspect, sarif_suffix: .archtest.sarif}
  java:
    - {module: gavel_tools_java, name: pmd, aspect: java_pmd_submission_aspect, sarif_suffix: .pmd.sarif, binary: pmd_binary}
    - {module: gavel_tools_java, name: cpd, aspect: java_cpd_submission_aspect, sarif_suffix: .cpd.sarif, binary: pmd_binary}
    - {module: gavel_tools_java, name: spotbugs, aspect: java_spotbugs_submission_aspect, sarif_suffix: .spotbugs.sarif, binary: spotbugs_binary}
    - {module: gavel_tools_java, name: error-prone, aspect: java_error_prone_submission_aspect, sarif_suffix: .errorprone.sarif, binary: error_prone_binary}
    - {module: gavel_tools_java, name: archtest, aspect: java_archtest_submission_aspect, sarif_suffix: .archtest.sarif}
  python:
    - {module: gavel_tools_python, name: ruff, aspect: python_ruff_submission_aspect, sarif_suffix: .ruff.sarif, binary: ruff_binary}
    - {module: gavel_tools_python, name: bandit, aspect: python_bandit_submission_aspect, sarif_suffix: .bandit.sarif, binary: bandit_binary}
    - {module: gavel_tools_python, name: pycompile, aspect: python_pycompile_submission_aspect, sarif_suffix: .pycompile.sarif}
    - {module: gavel_tools_python, name: archtest, aspect: python_archtest_submission_aspect, sarif_suffix: .archtest.sarif}
  typescript:
    - {module: gavel_tools_typescript, name: eslint, aspect: typescript_eslint_submission_aspect, sarif_suffix: .eslint.sarif}
    - {module: gavel_tools_typescript, name: archtest, aspect: typescript_archtest_submission_aspect, sarif_suffix: .archtest.sarif}
  rust:
    - name: clippy
      module: gavel_tools_rust
      aspect: rust_clippy_submission_aspect
      sarif_suffix: .clippy.sarif
      variants:
        - when_bazel_dep: rules_rs
          module: gavel_tools_rust_rs
          aspects_bzl: //lint/aspects:defs.bzl
          aspect: rust_clippy_submission_aspect
          build_flags: ["--@rules_rust//rust/settings:capture_clippy_output=True"]
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

func TestLintAspectsForLanguages_Go(t *testing.T) {
	aspects := catalog.LintAspectsForLanguages([]string{"go"}, nil)

	assert.Len(t, aspects, 1)
	assert.Equal(t, "go_golangci_lint_submission_aspect", aspects[0].Name)
	assert.Equal(t, "@gavel_tools_go//lint/aspects:defs.bzl%go_golangci_lint_submission_aspect", aspects[0].Path)
}

func TestLintAspectsForLanguages_GoGolangciCarriesExportStdlibFlag(t *testing.T) {
	aspects := catalog.LintAspectsForLanguages([]string{"go"}, nil)

	assert.Equal(t, []string{"--@rules_go//go/config:export_stdlib=True"}, aspects[0].BuildFlags,
		"golangci runs hermetic via the packages driver, which needs stdlib export data")
}

func TestLintAspectsForLanguages_NonGoAspectsHaveNoBuildFlags(t *testing.T) {
	for _, aspect := range catalog.LintAspectsForLanguages([]string{"java", "python", "rust", "typescript"}, nil) {
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

	aspects := catalog.LintAspectsForLanguages([]string{"java", "kotlin"}, nil)
	assert.Len(t, aspects, 1, "an aspect shared across languages is selected once")
}

func TestLintAspectsForLanguages_EmptyReturnsNothing(t *testing.T) {
	assert.Empty(t, catalog.LintAspectsForLanguages(nil, nil))
	assert.Empty(t, catalog.LintAspectsForLanguages([]string{}, nil))
}

func TestLintAspectsForLanguages_UnknownLanguage(t *testing.T) {
	assert.Empty(t, catalog.LintAspectsForLanguages([]string{"cobol"}, nil))
}

func TestSelectedAspects_ResolvesChosenToolsPerLanguage(t *testing.T) {
	aspects, err := catalog.SelectedAspects(map[string][]string{"go": {"golangci-lint", "archtest"}}, nil)

	require.NoError(t, err)
	require.Len(t, aspects, 2)
	assert.Equal(t, "go_golangci_lint_submission_aspect", aspects[0].Name)
	assert.Equal(t, "go_archtest_submission_aspect", aspects[1].Name)
	assert.Equal(t, []string{"--@rules_go//go/config:export_stdlib=True"}, aspects[0].BuildFlags)
}

func TestSelectedAspects_PreservesSelectionOrder(t *testing.T) {
	aspects, err := catalog.SelectedAspects(map[string][]string{"go": {"archtest", "golangci-lint"}}, nil)

	require.NoError(t, err)
	require.Len(t, aspects, 2)
	assert.Equal(t, "go_archtest_submission_aspect", aspects[0].Name, "the chosen order is preserved")
}

func TestSelectedAspects_OrdersLanguagesDeterministically(t *testing.T) {
	aspects, err := catalog.SelectedAspects(map[string][]string{"python": {"ruff"}, "go": {"golangci-lint"}}, nil)

	require.NoError(t, err)
	require.Len(t, aspects, 2)
	assert.Equal(t, "go_golangci_lint_submission_aspect", aspects[0].Name, "languages sorted alphabetically")
	assert.Equal(t, "python_ruff_submission_aspect", aspects[1].Name)
}

func TestSelectedAspects_ErrorsOnUnknownTool(t *testing.T) {
	_, err := catalog.SelectedAspects(map[string][]string{"go": {"clang-tidy"}}, nil)

	assert.ErrorContains(t, err, "clang-tidy")
	assert.ErrorContains(t, err, "go")
}

func TestSelectedAspects_ErrorsOnUnknownLanguage(t *testing.T) {
	_, err := catalog.SelectedAspects(map[string][]string{"cobol": {"anything"}}, nil)

	assert.ErrorContains(t, err, "anything")
}

func TestSelectedAspects_EmptySelectionReturnsNothing(t *testing.T) {
	aspects, err := catalog.SelectedAspects(map[string][]string{}, nil)

	require.NoError(t, err)
	assert.Empty(t, aspects)
}

func TestToolNamesForLanguage(t *testing.T) {
	assert.Equal(t, []string{"golangci-lint", "archtest"}, catalog.ToolNamesForLanguage("go"))
}

func TestToolNamesForLanguage_UnknownIsEmpty(t *testing.T) {
	assert.Empty(t, catalog.ToolNamesForLanguage("cobol"))
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
	aspects := catalog.ArchtestAspectsForLanguages([]string{"go"}, nil)

	assert.Len(t, aspects, 1)
	assert.Equal(t, "go_archtest_submission_aspect", aspects[0].Name)
}

func TestIsArchtestAspect(t *testing.T) {
	assert.True(t, catalog.IsArchtestAspect("go_archtest_submission_aspect"))
	assert.False(t, catalog.IsArchtestAspect("go_golangci_lint_submission_aspect"))
}

func TestAllAspectsForLanguages_Go(t *testing.T) {
	aspects := catalog.AllAspectsForLanguages([]string{"go"}, nil)

	assert.Len(t, aspects, 2)
	names := make([]string, len(aspects))
	for i, aspect := range aspects {
		names[i] = aspect.Name
	}
	assert.Contains(t, names, "go_golangci_lint_submission_aspect")
	assert.Contains(t, names, "go_archtest_submission_aspect")
}

func TestAllAspectsForLanguages_Java(t *testing.T) {
	assert.Len(t, catalog.AllAspectsForLanguages([]string{"java"}, nil), 5)
}

func TestAspectPaths_FormatsCommaSeparated(t *testing.T) {
	aspects := catalog.LintAspectsForLanguages([]string{"go"}, nil)

	assert.Equal(t, "@gavel_tools_go//lint/aspects:defs.bzl%go_golangci_lint_submission_aspect", catalog.AspectPaths(aspects))
}

func TestAspectPaths_MultipleAspects(t *testing.T) {
	got := catalog.AspectPaths(catalog.AllAspectsForLanguages([]string{"go"}, nil))

	assert.Contains(t, got, ",")
	assert.Contains(t, got, "go_golangci_lint_submission_aspect")
	assert.Contains(t, got, "go_archtest_submission_aspect")
}

func TestAspectPaths_EmptySlice(t *testing.T) {
	assert.Equal(t, "", catalog.AspectPaths(nil))
}

var rulesRs = catalog.Deps{"rules_rs": true}

func TestParseCatalog_RejectsVariantMissingRequiredField(t *testing.T) {
	cases := map[string]string{
		"when_bazel_dep": "{module: m, aspect: a}",
		"module":         "{when_bazel_dep: rules_rs, aspect: a}",
		"aspect":         "{when_bazel_dep: rules_rs, module: m}",
	}
	for missing, variant := range cases {
		t.Run(missing, func(t *testing.T) {
			yaml := "version: 1\naspects_bzl: //a:b\nlanguages:\n  rust:\n    - {name: clippy, aspect: x, variants: [" + variant + "]}\n"

			_, err := catalog.ParseCatalog([]byte(yaml))

			assert.ErrorContains(t, err, "variant")
			assert.ErrorContains(t, err, missing)
		})
	}
}

func TestSelectedAspects_WithoutMatchingDepUsesBaseEntry(t *testing.T) {
	for _, deps := range []catalog.Deps{nil, {}, {"rules_rust": true}} {
		aspects, err := catalog.SelectedAspects(map[string][]string{"rust": {"clippy"}}, deps)

		require.NoError(t, err)
		require.Len(t, aspects, 1)
		assert.Equal(t, "@gavel_tools_rust//lint/aspects:defs.bzl%rust_clippy_submission_aspect", aspects[0].Path)
		assert.Equal(t, "gavel_tools_rust", aspects[0].Module)
		assert.Empty(t, aspects[0].BuildFlags)
	}
}

func TestSelectedAspects_MatchingDepSelectsVariant(t *testing.T) {
	aspects, err := catalog.SelectedAspects(map[string][]string{"rust": {"clippy", "archtest"}}, rulesRs)

	require.NoError(t, err)
	require.Len(t, aspects, 2)
	assert.Equal(t, "rust_clippy_submission_aspect", aspects[0].Name)
	assert.Equal(t, "@gavel_tools_rust_rs//lint/aspects:defs.bzl%rust_clippy_submission_aspect", aspects[0].Path)
	assert.Equal(t, "gavel_tools_rust_rs", aspects[0].Module)
	assert.Equal(t, []string{"--@rules_rust//rust/settings:capture_clippy_output=True"}, aspects[0].BuildFlags)
	assert.Equal(t, "@gavel_tools//lint/aspects:defs.bzl%rust_archtest_submission_aspect", aspects[1].Path,
		"tools without variants are unaffected by consumer deps")
}

func TestLintAspectsForLanguages_MatchingDepSelectsVariant(t *testing.T) {
	aspects := catalog.LintAspectsForLanguages([]string{"rust"}, rulesRs)

	require.Len(t, aspects, 1)
	assert.Equal(t, "gavel_tools_rust_rs", aspects[0].Module)
}

func TestAspectFor_ResolvesSingleTool(t *testing.T) {
	aspect, ok := catalog.AspectFor("rust", "clippy", rulesRs)
	require.True(t, ok)
	assert.Equal(t, "gavel_tools_rust_rs", aspect.Module)

	base, ok := catalog.AspectFor("rust", "clippy", nil)
	require.True(t, ok)
	assert.Equal(t, "gavel_tools_rust", base.Module)

	_, ok = catalog.AspectFor("rust", "nope", nil)
	assert.False(t, ok)
}

func TestResolution_FirstMatchingVariantWinsAndFlagsFallBackToBase(t *testing.T) {
	parsed, err := catalog.ParseCatalog([]byte(`
version: 1
aspects_bzl: //lint/aspects:defs.bzl
languages:
  rust:
    - name: clippy
      aspect: base_aspect
      build_flags: ["--base"]
      variants:
        - {when_bazel_dep: first, module: m1, aspect: a1, build_flags: []}
        - {when_bazel_dep: second, module: m2, aspects_bzl: //x:y.bzl, aspect: a2}
`))
	require.NoError(t, err)
	catalog.SetCatalog(parsed)
	t.Cleanup(loadTestCatalog)

	both, _ := catalog.AspectFor("rust", "clippy", catalog.Deps{"first": true, "second": true})
	assert.Equal(t, "@m1//lint/aspects:defs.bzl%a1", both.Path, "variant aspects_bzl defaults to the catalog's")
	assert.Empty(t, both.BuildFlags, "an explicit empty build_flags replaces the base flags")

	second, _ := catalog.AspectFor("rust", "clippy", catalog.Deps{"second": true})
	assert.Equal(t, "@m2//x:y.bzl%a2", second.Path)
	assert.Equal(t, []string{"--base"}, second.BuildFlags, "absent build_flags keeps the base entry's")
}

func TestBaseEntryCarriesItsModule(t *testing.T) {
	aspects, err := catalog.SelectedAspects(map[string][]string{"java": {"pmd", "archtest"}}, nil)

	require.NoError(t, err)
	require.Len(t, aspects, 2)
	for _, aspect := range aspects {
		assert.Equal(t, "gavel_tools_java", aspect.Module)
		assert.Contains(t, aspect.Path, "@gavel_tools_java//lint/aspects:defs.bzl%")
	}
}

func TestToolWithoutModuleDefaultsToGavelTools(t *testing.T) {
	aspects, err := catalog.SelectedAspects(map[string][]string{"rust": {"archtest"}}, nil)

	require.NoError(t, err)
	require.Len(t, aspects, 1)
	assert.Equal(t, "gavel_tools", aspects[0].Module)
	assert.Equal(t, "@gavel_tools//lint/aspects:defs.bzl%rust_archtest_submission_aspect", aspects[0].Path)
}

func TestToolAspectsBzlOverridesCatalogDefault(t *testing.T) {
	parsed, err := catalog.ParseCatalog([]byte(`
version: 1
aspects_bzl: //lint/aspects:defs.bzl
languages:
  go:
    - {name: custom, module: m, aspects_bzl: //other:aspects.bzl, aspect: a, sarif_suffix: .x.sarif}
    - {name: plain, module: m, aspect: b, sarif_suffix: .y.sarif}
    - name: with_variant
      module: m
      aspects_bzl: //other:aspects.bzl
      aspect: c
      variants:
        - {when_bazel_dep: dep, module: v, aspect: c}
`))
	require.NoError(t, err)
	catalog.SetCatalog(parsed)
	t.Cleanup(loadTestCatalog)

	custom, _ := catalog.AspectFor("go", "custom", nil)
	plain, _ := catalog.AspectFor("go", "plain", nil)
	variant, _ := catalog.AspectFor("go", "with_variant", catalog.Deps{"dep": true})

	assert.Equal(t, "@m//other:aspects.bzl%a", custom.Path)
	assert.Equal(t, "@m//lint/aspects:defs.bzl%b", plain.Path)
	assert.Equal(t, "@v//other:aspects.bzl%c", variant.Path, "a variant without aspects_bzl inherits the tool's effective one")
}
