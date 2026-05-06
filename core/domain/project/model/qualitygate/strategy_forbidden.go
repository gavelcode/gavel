package qualitygate

import (
	"fmt"
	"strings"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/license"
)

type ForbiddenList struct {
	forbidden []string
}

func NewForbiddenList(forbidden []string) (ForbiddenList, error) {
	if len(forbidden) == 0 {
		return ForbiddenList{}, fmt.Errorf("%w: forbidden list must not be empty", ErrInvalidStrategy)
	}
	copied := make([]string, len(forbidden))
	copy(copied, forbidden)
	return ForbiddenList{forbidden: copied}, nil
}

func (f ForbiddenList) Forbidden() []string {
	copied := make([]string, len(f.forbidden))
	copy(copied, f.forbidden)
	return copied
}

func (f ForbiddenList) Equal(other Strategy) bool {
	otherForbiddenList, ok := other.(ForbiddenList)
	if !ok {
		return false
	}
	if len(f.forbidden) != len(otherForbiddenList.forbidden) {
		return false
	}
	set := make(map[string]struct{}, len(f.forbidden))
	for _, s := range f.forbidden {
		set[s] = struct{}{}
	}
	for _, s := range otherForbiddenList.forbidden {
		if _, found := set[s]; !found {
			return false
		}
	}
	return true
}

func (f ForbiddenList) Evaluate(content evidence.Content) Outcome {
	lc, ok := content.(license.Content)
	if !ok {
		return NewOutcome(true, "")
	}
	offending := f.findOffending(lc.Dependencies())
	if len(offending) == 0 {
		return NewOutcome(true, "")
	}
	return NewOutcome(false, fmt.Sprintf("forbidden licenses: %s", strings.Join(offending, ", ")))
}

func (f ForbiddenList) findOffending(deps []license.Dependency) []string {
	forbiddenSet := make(map[string]bool, len(f.forbidden))
	for _, lic := range f.forbidden {
		forbiddenSet[lic] = true
	}

	var offending []string
	for _, dep := range deps {
		if forbiddenSet[dep.License()] {
			offending = append(offending, fmt.Sprintf("%s (%s)", dep.Name(), dep.License()))
		}
	}
	return offending
}
