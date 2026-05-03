package license

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
)

type Content struct {
	dependencies []Dependency
}

func NewContent(dependencies []Dependency) (Content, error) {
	copied := make([]Dependency, len(dependencies))
	copy(copied, dependencies)

	return Content{
		dependencies: copied,
	}, nil
}

func (lc Content) Type() evidence.Type {
	return evidence.TypeSupplyChain
}

func (lc Content) Subtype() evidence.Subtype {
	return evidence.SubtypeLicense
}

func (lc Content) Dependencies() []Dependency {
	copied := make([]Dependency, len(lc.dependencies))
	copy(copied, lc.dependencies)
	return copied
}

func (lc Content) Merge(other evidence.Content) (evidence.Content, error) {
	otherLicenses, ok := other.(Content)
	if !ok {
		return nil, fmt.Errorf("%w: cannot merge license content with %T", ErrInvalidDependency, other)
	}
	merged := make([]Dependency, 0, len(lc.dependencies)+len(otherLicenses.dependencies))
	merged = append(merged, lc.dependencies...)
	merged = append(merged, otherLicenses.dependencies...)
	return NewContent(merged)
}
