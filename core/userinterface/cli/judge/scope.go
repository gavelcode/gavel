package judge

import "strings"

func scopeTargetsToPattern(targets []string, pattern string) []string {
	if !strings.HasSuffix(pattern, "...") {
		var exact []string
		for _, t := range targets {
			if t == pattern {
				exact = append(exact, t)
			}
		}
		return exact
	}

	prefix := strings.TrimSuffix(pattern, "...")
	var scoped []string
	for _, t := range targets {
		if strings.HasPrefix(t, prefix) {
			scoped = append(scoped, t)
		}
	}
	return scoped
}
