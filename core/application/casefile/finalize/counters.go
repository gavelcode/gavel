package finalize

type Counters struct {
	FindingsCount   int
	CoveragePercent float64
	NewCount        int
	ExistingCount   int
	ResolvedCount   int
	HasTracking     bool
}
