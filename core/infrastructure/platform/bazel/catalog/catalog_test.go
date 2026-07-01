package catalog_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/catalog"
)

func TestRuleKindsForLanguage_Go(t *testing.T) {
	kinds := catalog.RuleKindsForLanguage("go")
	assert.Contains(t, kinds, "go_library")
	assert.Contains(t, kinds, "go_binary")
	assert.Contains(t, kinds, "go_test")
}

func TestRuleKindsForLanguage_Java(t *testing.T) {
	kinds := catalog.RuleKindsForLanguage("java")
	assert.Contains(t, kinds, "java_library")
	assert.Contains(t, kinds, "java_binary")
	assert.Contains(t, kinds, "java_test")
}

func TestRuleKindsForLanguage_Rust(t *testing.T) {
	kinds := catalog.RuleKindsForLanguage("rust")
	assert.Contains(t, kinds, "rust_library")
	assert.Contains(t, kinds, "rust_binary")
	assert.Contains(t, kinds, "rust_test")
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
	for _, asp := range catalog.LintAspectsForLanguages([]string{"java", "python", "rust", "typescript"}) {
		assert.Empty(t, asp.BuildFlags, "%s must not force stdlib export — it is go-only and costly", asp.Name)
	}
}

func TestLintAspectsForLanguages_DeduplicatesSharedAspects(t *testing.T) {
	aspects := catalog.LintAspectsForLanguages([]string{"java", "kotlin"})

	assert.Len(t, aspects, 4)
}

func TestLintAspectsForLanguages_EmptyReturnsNothing(t *testing.T) {
	assert.Empty(t, catalog.LintAspectsForLanguages(nil))
	assert.Empty(t, catalog.LintAspectsForLanguages([]string{}))
}

func TestLintAspectsForLanguages_UnknownLanguage(t *testing.T) {
	aspects := catalog.LintAspectsForLanguages([]string{"cobol"})

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
	}, names)
}

func TestBinaryNames_EmptyReturnsNothing(t *testing.T) {
	assert.Empty(t, catalog.BinaryNames(nil))
	assert.Empty(t, catalog.BinaryNames([]string{}))
}

func TestBinaryNames_SingleLanguage(t *testing.T) {
	names := catalog.BinaryNames([]string{"go"})

	assert.Equal(t, []string{"golangci_lint_binary"}, names)
}

func TestBinaryNames_LanguageWithNoTools(t *testing.T) {
	names := catalog.BinaryNames([]string{"typescript"})

	assert.Empty(t, names)
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

func TestIsGavelExclusive(t *testing.T) {
	assert.True(t, catalog.IsGavelExclusive("java_error_prone_submission_aspect"))
	assert.True(t, catalog.IsGavelExclusive("java_cpd_submission_aspect"))
	assert.True(t, catalog.IsGavelExclusive("python_pycompile_submission_aspect"))
	assert.False(t, catalog.IsGavelExclusive("java_pmd_submission_aspect"))
	assert.False(t, catalog.IsGavelExclusive("go_golangci_lint_submission_aspect"))
}

func TestGavelExclusiveLintAspects_Java(t *testing.T) {
	aspects := catalog.GavelExclusiveLintAspects([]string{"java"})

	assert.Len(t, aspects, 2)
	names := make([]string, len(aspects))
	for i, a := range aspects {
		names[i] = a.Name
	}
	assert.Contains(t, names, "java_error_prone_submission_aspect")
	assert.Contains(t, names, "java_cpd_submission_aspect")
}

func TestGavelExclusiveLintAspects_Python(t *testing.T) {
	aspects := catalog.GavelExclusiveLintAspects([]string{"python"})

	assert.Len(t, aspects, 1)
	assert.Equal(t, "python_pycompile_submission_aspect", aspects[0].Name)
}

func TestGavelExclusiveLintAspects_Go(t *testing.T) {
	aspects := catalog.GavelExclusiveLintAspects([]string{"go"})

	assert.Empty(t, aspects)
}

func TestGavelExclusiveLintAspects_AllLanguages(t *testing.T) {
	aspects := catalog.GavelExclusiveLintAspects([]string{"go", "java", "python", "typescript", "rust"})

	assert.Len(t, aspects, 3)
}

func TestGavelExclusiveLintAspects_EmptyReturnsNothing(t *testing.T) {
	assert.Empty(t, catalog.GavelExclusiveLintAspects(nil))
}

func TestAllAspectsForLanguages_Go(t *testing.T) {
	aspects := catalog.AllAspectsForLanguages([]string{"go"})

	assert.Len(t, aspects, 2)
	names := make([]string, len(aspects))
	for i, a := range aspects {
		names[i] = a.Name
	}
	assert.Contains(t, names, "go_golangci_lint_submission_aspect")
	assert.Contains(t, names, "go_archtest_submission_aspect")
}

func TestAllAspectsForLanguages_Java(t *testing.T) {
	aspects := catalog.AllAspectsForLanguages([]string{"java"})

	assert.Len(t, aspects, 5)
}

func TestAllAspectsForLanguages_DeduplicatesSharedAspects(t *testing.T) {
	aspects := catalog.AllAspectsForLanguages([]string{"java", "kotlin"})

	assert.Len(t, aspects, 5)
}

func TestAspectPaths_FormatsCommaSeparated(t *testing.T) {
	aspects := catalog.LintAspectsForLanguages([]string{"go"})

	got := catalog.AspectPaths(aspects)

	assert.Equal(t, "@gavel//lint/aspects:defs.bzl%go_golangci_lint_submission_aspect", got)
}

func TestAspectPaths_MultipleAspects(t *testing.T) {
	aspects := catalog.AllAspectsForLanguages([]string{"go"})

	got := catalog.AspectPaths(aspects)

	assert.Contains(t, got, ",")
	assert.Contains(t, got, "go_golangci_lint_submission_aspect")
	assert.Contains(t, got, "go_archtest_submission_aspect")
}

func TestAspectPaths_EmptySlice(t *testing.T) {
	got := catalog.AspectPaths(nil)

	assert.Equal(t, "", got)
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

func TestSetModulePrefix_ChangesAspectPathsOutput(t *testing.T) {
	catalog.SetModulePrefix("@gavel_tools")
	t.Cleanup(func() { catalog.SetModulePrefix("@gavel") })

	aspects := catalog.LintAspectsForLanguages([]string{"go"})
	got := catalog.AspectPaths(aspects)

	assert.Equal(t, "@gavel_tools//lint/aspects:defs.bzl%go_golangci_lint_submission_aspect", got)
}
