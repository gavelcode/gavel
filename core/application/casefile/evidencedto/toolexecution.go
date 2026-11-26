package evidencedto

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/toolexecution"
)

type ToolExecution struct {
	Failures []ToolFailure
}

type ToolFailure struct {
	Tool     string
	Reason   string
	Degraded bool
	// Provenance is "catalog" or "bridge"; empty means catalog.
	Provenance string
}

func fromDomainToolExecution(content toolexecution.Content) ToolExecution {
	failures := content.Failures()
	out := make([]ToolFailure, 0, len(failures))
	for _, failed := range failures {
		out = append(out, ToolFailure{
			Tool:       failed.Tool(),
			Reason:     failed.Reason(),
			Degraded:   failed.Degraded(),
			Provenance: failed.Provenance().String(),
		})
	}
	return ToolExecution{Failures: out}
}

func toDomainToolExecution(in ToolExecution) (toolexecution.Content, error) {
	failures := make([]toolexecution.Failure, 0, len(in.Failures))
	for i, failed := range in.Failures {
		tf, err := newToolFailure(failed)
		if err != nil {
			return toolexecution.Content{}, fmt.Errorf("failures[%d]: %w", i, err)
		}
		failures = append(failures, tf)
	}
	return toolexecution.NewContent(failures)
}

func newToolFailure(failed ToolFailure) (toolexecution.Failure, error) {
	provenance, err := evidence.NewProvenance(failed.Provenance)
	if err != nil {
		return toolexecution.Failure{}, err
	}
	var domainFailure toolexecution.Failure
	if failed.Degraded {
		domainFailure, err = toolexecution.NewDegradedFailure(failed.Tool, failed.Reason)
	} else {
		domainFailure, err = toolexecution.NewFailure(failed.Tool, failed.Reason)
	}
	if err != nil {
		return toolexecution.Failure{}, err
	}
	return domainFailure.WithProvenance(provenance), nil
}

// ExtractToolFailures flattens the failures of every tool_execution evidence.
func ExtractToolFailures(evidences []Evidence) []ToolFailure {
	var out []ToolFailure
	for _, ev := range evidences {
		if ev.ToolExecution != nil {
			out = append(out, ev.ToolExecution.Failures...)
		}
	}
	return out
}
