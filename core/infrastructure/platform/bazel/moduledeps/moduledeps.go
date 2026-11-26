// Package moduledeps reads which bazel modules a consumer's root MODULE.bazel
// depends on, so gavel can pick the tool variant matching the rulesets the
// consumer already uses.
//
// Limitation: this is a textual scan, not a Starlark evaluation. It sees
// bazel_dep(...) calls written literally in MODULE.bazel and in files pulled in
// with include("//pkg:file"); deps produced by loops, variables or other
// constructs are not detected.
package moduledeps

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const rootModule = "MODULE.bazel"

var (
	bazelDepRe = regexp.MustCompile(`\bbazel_dep\s*\(([^)]*)\)`)
	nameRe     = regexp.MustCompile(`\bname\s*=\s*"([^"]+)"`)
	includeRe  = regexp.MustCompile(`\binclude\s*\(\s*"//([^":]*):?([^"]*)"`)
)

// Detect returns the bazel_dep module names declared by the workspace's root
// MODULE.bazel (and the files it includes). A missing file yields no deps.
func Detect(workspace string) map[string]bool {
	deps := make(map[string]bool)
	collect(workspace, rootModule, deps, make(map[string]bool))
	return deps
}

func collect(workspace, relative string, deps, visited map[string]bool) {
	if visited[relative] {
		return
	}
	visited[relative] = true

	data, err := os.ReadFile(filepath.Join(workspace, relative))
	if err != nil {
		return
	}
	content := stripComments(string(data))

	for _, call := range bazelDepRe.FindAllStringSubmatch(content, -1) {
		if match := nameRe.FindStringSubmatch(call[1]); match != nil {
			deps[match[1]] = true
		}
	}
	for _, include := range includeRe.FindAllStringSubmatch(content, -1) {
		collect(workspace, filepath.Join(include[1], include[2]), deps, visited)
	}
}

// stripComments drops '#' comments while leaving '#' inside string literals alone.
func stripComments(source string) string {
	var out strings.Builder
	inString := false
	inComment := false
	for _, r := range source {
		switch {
		case inComment:
			if r == '\n' {
				inComment = false
				out.WriteRune(r)
			}
		case r == '"':
			inString = !inString
			out.WriteRune(r)
		case r == '#' && !inString:
			inComment = true
		default:
			out.WriteRune(r)
		}
	}
	return out.String()
}
