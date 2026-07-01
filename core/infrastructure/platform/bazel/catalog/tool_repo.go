package catalog

// languageBinaries maps a language to the linter tool binary repo-rule names
// it relies on. The repos themselves (and their versions) are owned by the
// gavel_tools module; this list only drives `gavel init` informational output.
var languageBinaries = map[string][]string{
	"go":     {"golangci_lint_binary"},
	"java":   {"pmd_binary", "spotbugs_binary", "error_prone_binary"},
	"python": {"ruff_binary", "bandit_binary"},
}

func BinaryNames(languages []string) []string {
	seen := make(map[string]bool)
	var names []string
	for _, lang := range languages {
		if seen[lang] {
			continue
		}
		seen[lang] = true
		names = append(names, languageBinaries[lang]...)
	}
	return names
}
