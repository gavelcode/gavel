package collectevidence

import (
	"github.com/usegavel/gavel/core/application/casefile/classifyarch"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
)

type Result struct {
	Evidences       []evidencedto.Evidence
	FindingsCount   int
	ViolationsCount int
	CovPercent      float64
	NCCPercent      float64
	CoverageByFile  []evidencedto.FileCoverage
	RawSARIF        []RawFile
	RawLCOV         []byte
	SARIFDocs       [][]byte
	Findings        []evidencedto.Finding
	Fingerprints    []string
	Violations      []evidencedto.Violation
	ArchIDs         []string
	ArchDelta       classifyarch.Result
	BuildWarning    string
}

type RawFile struct {
	Format string
	Source string
	Data   []byte
}
