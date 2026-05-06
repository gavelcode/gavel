package archpolicy

import (
	"fmt"
	"strings"
)

type DenyRule struct {
	name   string
	source string
	deny   []string
}

func NewDenyRule(name, source string, deny []string) (DenyRule, error) {
	if strings.TrimSpace(name) == "" {
		return DenyRule{}, fmt.Errorf("%w: name must not be empty", ErrInvalidDenyRule)
	}
	if strings.TrimSpace(source) == "" {
		return DenyRule{}, fmt.Errorf("%w: source must not be empty", ErrInvalidDenyRule)
	}
	if len(deny) == 0 {
		return DenyRule{}, fmt.Errorf("%w: at least one deny target required", ErrInvalidDenyRule)
	}
	for i, d := range deny {
		if strings.TrimSpace(d) == "" {
			return DenyRule{}, fmt.Errorf("%w: deny[%d] must not be empty", ErrInvalidDenyRule, i)
		}
	}
	copied := make([]string, len(deny))
	copy(copied, deny)
	return DenyRule{name: name, source: source, deny: copied}, nil
}

func (r DenyRule) Name() string {
	return r.name
}

func (r DenyRule) Source() string {
	return r.source
}

func (r DenyRule) Deny() []string {
	copied := make([]string, len(r.deny))
	copy(copied, r.deny)
	return copied
}
