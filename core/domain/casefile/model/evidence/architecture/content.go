package architecture

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
)

type Content struct {
	violations []Violation
}

func NewContent(violations []Violation) (Content, error) {
	copied := make([]Violation, len(violations))
	copy(copied, violations)

	return Content{
		violations: copied,
	}, nil
}

func (ac Content) Type() evidence.Type {
	return evidence.TypeArchitecture
}

func (ac Content) Subtype() evidence.Subtype {
	return evidence.SubtypeArchitecture
}

func (ac Content) Violations() []Violation {
	copied := make([]Violation, len(ac.violations))
	copy(copied, ac.violations)
	return copied
}

func (ac Content) Merge(other evidence.Content) (evidence.Content, error) {
	otherArchitecture, ok := other.(Content)
	if !ok {
		return nil, fmt.Errorf("%w: cannot merge architecture content with %T", ErrInvalidViolation, other)
	}
	merged := make([]Violation, 0, len(ac.violations)+len(otherArchitecture.violations))
	merged = append(merged, ac.violations...)
	merged = append(merged, otherArchitecture.violations...)
	return NewContent(merged)
}
