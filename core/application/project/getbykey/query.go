package getbykey

import (
	"fmt"
	"strings"
)

type Query struct {
	key string
}

func NewQuery(key string) (Query, error) {
	if strings.TrimSpace(key) == "" {
		return Query{}, fmt.Errorf("%w: key must not be empty", ErrInvalidQuery)
	}
	return Query{key: key}, nil
}

func (q Query) Key() string { return q.key }
