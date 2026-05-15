package getbaseline

import (
	"fmt"
	"strings"
)

type Query struct {
	key    string
	branch string
}

func NewQuery(key, branch string) (Query, error) {
	if strings.TrimSpace(key) == "" {
		return Query{}, fmt.Errorf("%w: key must not be empty", ErrInvalidQuery)
	}
	return Query{key: key, branch: branch}, nil
}

func (q Query) Key() string    { return q.key }
func (q Query) Branch() string { return q.branch }
