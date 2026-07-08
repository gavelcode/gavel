package getbykey

import (
	"fmt"
	"strings"
)

type Query struct {
	tenantID string
	key      string
}

func NewQuery(tenantID, key string) (Query, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Query{}, fmt.Errorf("%w: tenantID must not be empty", ErrInvalidQuery)
	}
	if strings.TrimSpace(key) == "" {
		return Query{}, fmt.Errorf("%w: key must not be empty", ErrInvalidQuery)
	}
	return Query{tenantID: tenantID, key: key}, nil
}

func (q Query) TenantID() string { return q.tenantID }
func (q Query) Key() string      { return q.key }
