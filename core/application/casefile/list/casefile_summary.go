package list

import "time"

type CaseFileSummary struct {
	ID               string
	ProjectID        string
	CommitSHA        string
	Branch           string
	StartedAt        time.Time
	VerdictOutcome   string
	TotalFindings    int
	NewFindings      int
	ExistingFindings int
	ResolvedFindings int
	CoveragePercent  *float64
	CreatedAt        time.Time
}
