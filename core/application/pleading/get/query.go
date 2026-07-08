package get

import (
	"fmt"
	"strings"
)

type Query struct {
	tenantID string
	id       string
}

func NewQuery(tenantID, pleadingID string) (Query, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Query{}, fmt.Errorf("%w: tenantID must not be empty", ErrInvalidQuery)
	}
	if strings.TrimSpace(pleadingID) == "" {
		return Query{}, fmt.Errorf("%w: id must not be empty", ErrInvalidQuery)
	}
	return Query{tenantID: tenantID, id: pleadingID}, nil
}

func (q Query) TenantID() string { return q.tenantID }
func (q Query) ID() string       { return q.id }
