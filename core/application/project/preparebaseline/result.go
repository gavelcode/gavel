package preparebaseline

type ProjectBaseline struct {
	ProjectName      string
	FingerprintCount int
	ArchIDCount      int
	HasPrevious      bool
	Source           string
}

type Result struct {
	Baselines []ProjectBaseline
}
