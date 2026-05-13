package finalize

type Delta struct {
	NewCount        int
	FixedCount      int
	ExistingCount   int
	NewFingerprints map[string]bool
	HasPrevious     bool

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
