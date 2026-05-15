package search

import (
	"fmt"
	"strings"
)

type Query struct {
	text  string
	limit int
}

func NewQuery(text string, limit int) (Query, error) {
	if strings.TrimSpace(text) == "" {
		return Query{}, fmt.Errorf("%w: text must not be empty", ErrInvalidQuery)
	}
	if limit <= 0 {
		return Query{}, fmt.Errorf("%w: limit must be positive", ErrInvalidQuery)
	}
	return Query{text: text, limit: limit}, nil
}

func (q Query) Text() string { return q.text }
func (q Query) Limit() int   { return q.limit }
