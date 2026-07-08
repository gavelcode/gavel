package getbaseline

import (
	"fmt"
	"strings"
)

type Query struct {
	tenantID string
	key      string
	branch   string
}

func NewQuery(tenantID, key, branch string) (Query, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Query{}, fmt.Errorf("%w: tenantID must not be empty", ErrInvalidQuery)
	}
	if strings.TrimSpace(key) == "" {
		return Query{}, fmt.Errorf("%w: key must not be empty", ErrInvalidQuery)
	}
	return Query{tenantID: tenantID, key: key, branch: branch}, nil
}

func (q Query) TenantID() string { return q.tenantID }
func (q Query) Key() string      { return q.key }
func (q Query) Branch() string   { return q.branch }
