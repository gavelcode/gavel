package list

import "fmt"

type Query struct {
	limit  int
	offset int
}

func NewQuery(limit, offset int) (Query, error) {
	if limit <= 0 {
		return Query{}, fmt.Errorf("%w: limit must be positive", ErrInvalidQuery)
	}
	if offset < 0 {
		return Query{}, fmt.Errorf("%w: offset must not be negative", ErrInvalidQuery)
	}
	return Query{limit: limit, offset: offset}, nil
}

func (q Query) Limit() int  { return q.limit }
func (q Query) Offset() int { return q.offset }
