package classifyarch

type Result struct {
	NewCount      int
	FixedCount    int
	ExistingCount int
	NewIDs        map[string]bool
}
