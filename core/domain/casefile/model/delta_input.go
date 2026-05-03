package model

type DeltaInput struct {
	FindingsResolved int
	ArchResolved     int
	PreviousCoverage *float64
	CurrentCoverage  float64
}
