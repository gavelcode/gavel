package judge

import "strings"

func scopeTargetsToPattern(targets []string, pattern string, excludes []string) []string {
	if !strings.HasSuffix(pattern, "...") {
		var exact []string
		for _, t := range targets {
			if t == pattern && !isExcluded(t, excludes) {
				exact = append(exact, t)
			}
		}
		return exact
	}

	prefix := strings.TrimSuffix(pattern, "...")
	var scoped []string
	for _, t := range targets {
		if strings.HasPrefix(t, prefix) && !isExcluded(t, excludes) {
			scoped = append(scoped, t)
		}
	}
	return scoped
}

func isExcluded(target string, excludes []string) bool {
	pkg := target
	if i := strings.LastIndex(target, ":"); i >= 0 {
		pkg = target[:i]
	}
	for _, exclude := range excludes {
		if !strings.HasSuffix(exclude, "...") {
			if target == exclude {
				return true
			}
			continue
		}
		base := strings.TrimSuffix(strings.TrimSuffix(exclude, "..."), "/")
		if pkg == base || strings.HasPrefix(pkg, base+"/") {
			return true
		}
	}
	return false
}
