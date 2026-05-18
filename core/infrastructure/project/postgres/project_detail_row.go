package postgres

import "time"

type projectDetailRow struct {
	id, key, name, targetPattern, defaultBranch string
	latestVerdict                               string
	totalFindings                               int
	createdAt                                   time.Time
	languages                                   []string
	qgRules                                     []qgRuleRow
	severityCounts                              map[string]int
}
