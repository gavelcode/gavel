package evidencedto

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/tracking"
)

type Tracking struct {
	NewFindings      []Finding
	ExistingFindings []Finding
	ResolvedCount    int
}

func TrackingFromDomain(tr tracking.Result) Tracking {
	return Tracking{
		NewFindings:      fromDomainFindings(tr.NewFindings()),
		ExistingFindings: fromDomainFindings(tr.ExistingFindings()),
		ResolvedCount:    tr.ResolvedCount(),
	}
}

func TrackingToDomain(trackingDTO Tracking) (tracking.Result, error) {
	newFindings, err := toDomainFindings(trackingDTO.NewFindings)
	if err != nil {
		return tracking.Result{}, fmt.Errorf("new findings: %w", err)
	}
	existingFindings, err := toDomainFindings(trackingDTO.ExistingFindings)
	if err != nil {
		return tracking.Result{}, fmt.Errorf("existing findings: %w", err)
	}
	return tracking.NewResult(newFindings, existingFindings, trackingDTO.ResolvedCount), nil
}
