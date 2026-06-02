package get

import (
	"fmt"
	"strings"
)

type Query struct {
	id string
}

func NewQuery(id string) (Query, error) {
	if strings.TrimSpace(id) == "" {
		return Query{}, fmt.Errorf("%w: id must not be empty", ErrInvalidQuery)
	}
	return Query{id: id}, nil
}

func (q Query) ID() string { return q.id }
