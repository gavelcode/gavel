package get

import "time"

type EvidenceSummary struct {
	ID          string
	Subtype     string
	Source      string
	CollectedAt time.Time
}
