package get

import (
	"fmt"
	"strings"
)

type Query struct {
	name string
}

func NewQuery(name string) (Query, error) {
	if strings.TrimSpace(name) == "" {
		return Query{}, fmt.Errorf("%w: name must not be empty", ErrInvalidQuery)
	}
	return Query{name: name}, nil
}

func (q Query) Name() string { return q.name }
