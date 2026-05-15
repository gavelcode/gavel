package list

import "time"

type ProjectSummary struct {
	ID            string
	Key           string
	Name          string
	DefaultBranch string
	LatestVerdict string
	TotalFindings int
	CreatedAt     time.Time
}
