package evidencedto

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/toolexecution"
)

type ToolExecution struct {
	Failures []ToolFailure
}

type ToolFailure struct {
	Tool     string
	Reason   string
	Degraded bool
}

func fromDomainToolExecution(content toolexecution.Content) ToolExecution {
	failures := content.Failures()
	out := make([]ToolFailure, 0, len(failures))
	for _, failed := range failures {
		out = append(out, ToolFailure{
			Tool:     failed.Tool(),
			Reason:   failed.Reason(),
			Degraded: failed.Degraded(),
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
	if failed.Degraded {
		return toolexecution.NewDegradedFailure(failed.Tool, failed.Reason)
	}
	return toolexecution.NewFailure(failed.Tool, failed.Reason)
}
