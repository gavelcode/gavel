package catalog

var languageRuleKinds = map[string][]string{
	"go":         {"go_library", "go_binary", "go_test"},
	"java":       {"java_library", "java_binary", "java_test"},
	"kotlin":     {"kt_jvm_library", "kt_jvm_binary", "kt_jvm_test", "java_library", "java_binary", "java_test"},
	"python":     {"py_library", "py_binary", "py_test"},
	"typescript": {"ts_project", "ts_library", "js_library", "js_binary", "js_test"},
	"rust":       {"rust_library", "rust_binary", "rust_test"},
}

func RuleKindsForLanguage(language string) []string {
	kinds, ok := languageRuleKinds[language]
	if !ok {
		return nil
	}
	out := make([]string, len(kinds))
	copy(out, kinds)
	return out
}
