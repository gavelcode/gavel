package catalog

import "strings"

const archtestAspectSuffix = "_archtest_submission_aspect"

type Aspect struct {
	Name        string
	Path        string
	SARIFSuffix string
	BuildFlags  []string
	// Module is the bazel module that publishes the aspect (effective, after variant resolution).
	Module string
}

func IsArchtestAspect(name string) bool {
	return strings.HasSuffix(name, archtestAspectSuffix)
}

func AllAspectsForLanguages(languages []string, deps Deps) []Aspect {
	return active().aspects(languages, deps, func(Tool) bool { return true })
}

func LintAspectsForLanguages(languages []string, deps Deps) []Aspect {
	return active().aspects(languages, deps, func(tool Tool) bool { return !IsArchtestAspect(tool.Aspect) })
}

func ArchtestAspectsForLanguages(languages []string, deps Deps) []Aspect {
	return active().aspects(languages, deps, func(tool Tool) bool { return IsArchtestAspect(tool.Aspect) })
}

func AspectPaths(aspects []Aspect) string {
	paths := make([]string, len(aspects))
	for i, aspect := range aspects {
		paths[i] = aspect.Path
	}
	return strings.Join(paths, ",")
}

func AspectNames(languages []string) []string {
	return active().aspectNames(languages)
}

func SelectedAspects(selection map[string][]string, deps Deps) ([]Aspect, error) {
	return active().selectedAspects(selection, deps)
}

// AspectFor resolves one tool against the consumer's deps, as init and judge both do.
func AspectFor(language, tool string, deps Deps) (Aspect, bool) {
	return active().aspectForTool(language, tool, deps)
}

func ToolNamesForLanguage(language string) []string {
	return active().toolNames(language)
}
