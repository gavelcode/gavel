package toolexecution

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
)

type Content struct {
	failures []Failure
}

func NewContent(failures []Failure) (Content, error) {
	copied := make([]Failure, len(failures))
	copy(copied, failures)

	return Content{
		failures: copied,
	}, nil
}

func (c Content) Type() evidence.Type {
	return evidence.TypeAnalysis
}

func (c Content) Subtype() evidence.Subtype {
	return evidence.SubtypeToolExecution
}

func (c Content) Failures() []Failure {
	copied := make([]Failure, len(c.failures))
	copy(copied, c.failures)
	return copied
}

func (c Content) Merge(other evidence.Content) (evidence.Content, error) {
	otherContent, ok := other.(Content)
	if !ok {
		return nil, fmt.Errorf("%w: cannot merge tool_execution content with %T", ErrInvalidFailure, other)
	}
	merged := make([]Failure, 0, len(c.failures)+len(otherContent.failures))
	merged = append(merged, c.failures...)
	merged = append(merged, otherContent.failures...)
	return NewContent(merged)
}
