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
	BuildFlags  []string
}

func IsArchtestAspect(name string) bool {
	return strings.HasSuffix(name, archtestAspectSuffix)
}

func AllAspectsForLanguages(languages []string) []Aspect {
	return active().aspects(languages, func(Tool) bool { return true })
}

func LintAspectsForLanguages(languages []string) []Aspect {
	return active().aspects(languages, func(tool Tool) bool { return !IsArchtestAspect(tool.Aspect) })
}

func ArchtestAspectsForLanguages(languages []string) []Aspect {
	return active().aspects(languages, func(tool Tool) bool { return IsArchtestAspect(tool.Aspect) })
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

func SelectedAspects(selection map[string][]string) ([]Aspect, error) {
	return active().selectedAspects(selection)
}

func ToolNamesForLanguage(language string) []string {
	return active().toolNames(language)
}
