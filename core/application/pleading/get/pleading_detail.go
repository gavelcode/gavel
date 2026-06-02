package get

import "time"

type PleadingDetail struct {
	ID           string
	ProjectID    string
	Number       int
	Title        string
	Petitioner   string
	SourceBranch string
	TargetBranch string
	CommitSHA    string
	Status       string
	GateResult   *GateResult
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
