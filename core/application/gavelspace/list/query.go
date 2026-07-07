package list

import (
	"fmt"
	"strings"
)

type Query struct {
	tenantID string
	limit    int
	offset   int
}

func NewQuery(tenantID string, limit, offset int) (Query, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Query{}, fmt.Errorf("%w: tenantID must not be empty", ErrInvalidQuery)
	}
	if limit <= 0 {
		return Query{}, fmt.Errorf("%w: limit must be positive", ErrInvalidQuery)
	}
	if offset < 0 {
		return Query{}, fmt.Errorf("%w: offset must not be negative", ErrInvalidQuery)
	}
	return Query{tenantID: tenantID, limit: limit, offset: offset}, nil
}

func (q Query) TenantID() string { return q.tenantID }
func (q Query) Limit() int       { return q.limit }
func (q Query) Offset() int      { return q.offset }
