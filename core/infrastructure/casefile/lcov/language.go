package lcov

import "path/filepath"

const fallbackLanguage = "other"

var extensionToLanguage = map[string]string{
	".go":    "go",
	".java":  "java",
	".kt":    "java",
	".py":    "python",
	".ts":    "typescript",
	".tsx":   "typescript",
	".js":    "javascript",
	".jsx":   "javascript",
	".rs":    "rust",
	".rb":    "ruby",
	".cpp":   "cpp",
	".cc":    "cpp",
	".c":     "c",
	".h":     "c",
	".cs":    "csharp",
	".swift": "swift",
}

func languageFromPath(path string) string {
	ext := filepath.Ext(path)
	if lang, ok := extensionToLanguage[ext]; ok {
		return lang
	}
	return fallbackLanguage
}
