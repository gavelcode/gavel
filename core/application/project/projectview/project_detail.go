package projectview

import "time"

type ProjectDetail struct {
	ID               string
	Key              string
	Name             string
	DefaultBranch    string
	LatestVerdict    string
	TotalFindings    int
	CreatedAt        time.Time
	TargetPattern    string
	Languages        []string
	QualityGateRules []QualityGateRuleView
	SeverityCounts   map[string]int
}
