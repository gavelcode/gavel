package finding

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
)

type Content struct {
	subtype  evidence.Subtype
	findings []Finding
}

func NewContent(subtype evidence.Subtype, findings []Finding) (Content, error) {
	if !evidence.IsSubtypeFindingBased(subtype) {
		return Content{}, fmt.Errorf("%w: subtype %q is not finding-based", ErrInvalidContent, subtype)
	}

	copied := make([]Finding, len(findings))
	copy(copied, findings)

	return Content{
		subtype:  subtype,
		findings: copied,
	}, nil
}

func (fc Content) Type() evidence.Type {
	return fc.subtype.Type()
}

func (fc Content) Subtype() evidence.Subtype {
	return fc.subtype
}

func (fc Content) Findings() []Finding {
	copied := make([]Finding, len(fc.findings))
	copy(copied, fc.findings)
	return copied
}

func (fc Content) Merge(other evidence.Content) (evidence.Content, error) {
	otherFindings, ok := other.(Content)
	if !ok {
		return nil, fmt.Errorf("%w: cannot merge finding content with %T", ErrInvalidContent, other)
	}
	if fc.subtype != otherFindings.subtype {
		return nil, fmt.Errorf("%w: cannot merge findings with subtypes %q and %q", ErrInvalidContent, fc.subtype, otherFindings.subtype)
	}
	merged := make([]Finding, 0, len(fc.findings)+len(otherFindings.findings))
	merged = append(merged, fc.findings...)
	merged = append(merged, otherFindings.findings...)
	return NewContent(fc.subtype, merged)
}
