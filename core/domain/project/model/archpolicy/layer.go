package archpolicy

import (
	"fmt"
	"strings"
)

type Layer struct {
	name     string
	patterns []string
}

func NewLayer(name string, patterns []string) (Layer, error) {
	if strings.TrimSpace(name) == "" {
		return Layer{}, fmt.Errorf("%w: name must not be empty", ErrInvalidLayer)
	}
	if len(patterns) == 0 {
		return Layer{}, fmt.Errorf("%w: at least one pattern required", ErrInvalidLayer)
	}
	for i, p := range patterns {
		if strings.TrimSpace(p) == "" {
			return Layer{}, fmt.Errorf("%w: pattern[%d] must not be empty", ErrInvalidLayer, i)
		}
	}
	copied := make([]string, len(patterns))
	copy(copied, patterns)
	return Layer{name: name, patterns: copied}, nil
}

func (l Layer) Name() string {
	return l.name
}

func (l Layer) Patterns() []string {
	copied := make([]string, len(l.patterns))
	copy(copied, l.patterns)
	return copied
}
