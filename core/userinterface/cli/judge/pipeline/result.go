package pipeline

import (
	"time"

	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	corejudge "github.com/usegavel/gavel/core/application/casefile/judge"
)

type Result struct {
	Name                   string
	Verdict                string
	CommitSHA              string
	Branch                 string
	StartedAt              time.Time
	FindingsCount          int
	ViolationsCount        int
	CoveragePercent        float64
	CoverageSkipped        bool
	NewCodeCoveragePercent float64
	CoverageByFile         []evidencedto.FileCoverage
	PreviousCoverageByFile []evidencedto.FileCoverage
	Rulings                []corejudge.RulingView
	Findings               []evidencedto.Finding
	Violations             []evidencedto.Violation
	Delta                  Delta
	FirstRun               bool
	RawSARIFDocs           [][]byte
	ServerFailed           bool
	BuildWarning           string
}
