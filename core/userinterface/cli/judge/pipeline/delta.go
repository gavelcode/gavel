package pipeline

type Delta struct {
	NewCount        int
	FixedCount      int
	ExistingCount   int
	NewFingerprints map[string]bool
	FindingsDelta   int
	CoverageDelta   float64
	ViolationsDelta int
	HasPrevious     bool

	NewViolationsCount      int
	FixedViolationsCount    int
	ExistingViolationsCount int
	NewViolationIDs         map[string]bool
	HasArchPrevious         bool
}
