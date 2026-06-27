package finalize

import "github.com/usegavel/gavel/core/application/casefile/evidencedto"

type Delta struct {
	NewCount        int
	FixedCount      int
	ExistingCount   int
	NewFingerprints map[string]bool
	HasPrevious     bool

	PreviousCoveragePercent *float64
	PreviousFileCoverage    []evidencedto.FileCoverage

	NewViolationsCount      int
	FixedViolationsCount    int
	ExistingViolationsCount int
	NewViolationIDs         map[string]bool
	HasArchPrevious         bool
}

type ArchDeltaInput struct {
	NewCount      int
	FixedCount    int
	ExistingCount int
	NewIDs        map[string]bool
}
