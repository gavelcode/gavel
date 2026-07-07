package get

import (
	"fmt"
	"strings"
)

type Query struct {
	tenantID string
	name     string
}

func NewQuery(tenantID, name string) (Query, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Query{}, fmt.Errorf("%w: tenantID must not be empty", ErrInvalidQuery)
	}
	if strings.TrimSpace(name) == "" {
		return Query{}, fmt.Errorf("%w: name must not be empty", ErrInvalidQuery)
	}
	return Query{tenantID: tenantID, name: name}, nil
}

func (q Query) TenantID() string { return q.tenantID }
func (q Query) Name() string     { return q.name }
