package qualitygate

import (
	"fmt"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
)

type Gate struct {
	rules []Rule
}

func NewGate(rules []Rule) (Gate, error) {
	if err := validateNoDuplicateSubtypes(rules); err != nil {
		return Gate{}, err
	}
	copied := make([]Rule, len(rules))
	copy(copied, rules)
	return Gate{rules: copied}, nil
}

func validateNoDuplicateSubtypes(rules []Rule) error {
	seen := make(map[evidence.Subtype]bool, len(rules))
	for _, r := range rules {
		if seen[r.Subtype()] {
			return fmt.Errorf("%w: duplicate subtype %q", ErrInvalidGate, r.Subtype())
		}
		seen[r.Subtype()] = true
	}
	return nil
}

func (g Gate) Rules() []Rule {
	copied := make([]Rule, len(g.rules))
	copy(copied, g.rules)
	return copied
}

func (g Gate) RuleForSubtype(subtype evidence.Subtype) (Rule, bool) {
	for _, r := range g.rules {
		if r.Subtype() == subtype {
			return r, true
		}
	}
	return Rule{}, false
}

func (g Gate) Equal(other Gate) bool {
	if len(g.rules) != len(other.rules) {
		return false
	}
	byType := make(map[evidence.Subtype]Rule, len(g.rules))
	for _, r := range g.rules {
		byType[r.Subtype()] = r
	}
	for _, or := range other.rules {
		gr, ok := byType[or.Subtype()]
		if !ok {
			return false
		}
		if !gr.Equal(or) {
			return false
		}
	}
	return true
}
