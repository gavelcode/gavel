package catalog

import "strings"

var modulePrefix = "@gavel"

func ModulePrefix() string {
	return modulePrefix
}

func SetModulePrefix(prefix string) {
	modulePrefix = prefix
}

const archtestAspectSuffix = "_archtest_submission_aspect"

type Aspect struct {
	Name        string
	Path        string
	SARIFSuffix string
}

var languageAspects = map[string][]string{
	"go":         {"go_golangci_lint_submission_aspect", "go_archtest_submission_aspect"},
	"java":       {"java_pmd_submission_aspect", "java_cpd_submission_aspect", "java_spotbugs_submission_aspect", "java_error_prone_submission_aspect", "java_archtest_submission_aspect"},
	"kotlin":     {"java_pmd_submission_aspect", "java_cpd_submission_aspect", "java_spotbugs_submission_aspect", "java_error_prone_submission_aspect", "java_archtest_submission_aspect"},
	"python":     {"python_pycompile_submission_aspect", "python_ruff_submission_aspect", "python_bandit_submission_aspect", "python_archtest_submission_aspect"},
	"typescript": {"typescript_eslint_submission_aspect", "typescript_archtest_submission_aspect"},
	"rust":       {"rust_clippy_submission_aspect", "rust_archtest_submission_aspect"},
}

var gavelExclusiveAspects = map[string]bool{
	"java_error_prone_submission_aspect": true,
	"java_cpd_submission_aspect":         true,
	"python_pycompile_submission_aspect": true,
}

var aspectSARIFSuffix = map[string]string{
	"go_golangci_lint_submission_aspect":    ".golangci.sarif",
	"go_archtest_submission_aspect":         ".archtest.sarif",
	"java_pmd_submission_aspect":            ".pmd.sarif",
	"java_cpd_submission_aspect":            ".cpd.sarif",
	"java_spotbugs_submission_aspect":       ".spotbugs.sarif",
	"java_error_prone_submission_aspect":    ".errorprone.sarif",
	"java_archtest_submission_aspect":       ".archtest.sarif",
	"python_pycompile_submission_aspect":    ".pycompile.sarif",
	"python_ruff_submission_aspect":         ".ruff.sarif",
	"python_bandit_submission_aspect":       ".bandit.sarif",
	"python_archtest_submission_aspect":     ".archtest.sarif",
	"typescript_eslint_submission_aspect":   ".eslint.sarif",
	"typescript_archtest_submission_aspect": ".archtest.sarif",
	"rust_clippy_submission_aspect":         ".clippy.sarif",
	"rust_archtest_submission_aspect":       ".archtest.sarif",
}

var defaultLanguages = []string{"go", "java", "python", "typescript", "rust"}

func IsArchtestAspect(name string) bool {
	return strings.HasSuffix(name, archtestAspectSuffix)
}

func IsGavelExclusive(name string) bool {
	return gavelExclusiveAspects[name]
}

func GavelExclusiveLintAspects(languages []string) []Aspect {
	return filterAspects(languages, func(name string) bool {
		return !IsArchtestAspect(name) && IsGavelExclusive(name)
	})
}

func AllAspectsForLanguages(languages []string) []Aspect {
	return filterAspects(languages, func(_ string) bool { return true })
}

func LintAspectsForLanguages(languages []string) []Aspect {
	return filterAspects(languages, func(name string) bool { return !IsArchtestAspect(name) })
}

func ArchtestAspectsForLanguages(languages []string) []Aspect {
	return filterAspects(languages, IsArchtestAspect)
}

func filterAspects(languages []string, keep func(string) bool) []Aspect {
	if len(languages) == 0 {
		languages = defaultLanguages
	}
	seen := make(map[string]bool)
	var aspects []Aspect
	for _, lang := range languages {
		for _, name := range languageAspects[lang] {
			if seen[name] || !keep(name) {
				continue
			}
			seen[name] = true
			aspects = append(aspects, Aspect{
				Name:        name,
				Path:        modulePrefix + "//lint/aspects:defs.bzl%" + name,
				SARIFSuffix: aspectSARIFSuffix[name],
			})
		}
	}
	return aspects
}

func AspectPaths(aspects []Aspect) string {
	paths := make([]string, len(aspects))
	for i, a := range aspects {
		paths[i] = a.Path
	}
	return strings.Join(paths, ",")
}

func AspectNames(languages []string) []string {
	if len(languages) == 0 {
		languages = defaultLanguages
	}
	var names []string
	for _, lang := range languages {
		names = append(names, languageAspects[lang]...)
	}
	return names
}
