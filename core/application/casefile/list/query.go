package list

import (
	"fmt"
	"strings"
)

type Query struct {
	projectID  string
	gavelspace string
	limit      int
	offset     int
}

func NewQuery(projectID, gavelspace string, limit, offset int) (Query, error) {
	if strings.TrimSpace(projectID) == "" && strings.TrimSpace(gavelspace) == "" {
		return Query{}, fmt.Errorf("%w: projectID or gavelspace must be set", ErrInvalidQuery)
	}
	if limit <= 0 {
		return Query{}, fmt.Errorf("%w: limit must be greater than zero", ErrInvalidQuery)
	}
	if offset < 0 {
		return Query{}, fmt.Errorf("%w: offset must not be negative", ErrInvalidQuery)
	}
	return Query{projectID: projectID, gavelspace: gavelspace, limit: limit, offset: offset}, nil
}

func (q Query) ProjectID() string  { return q.projectID }
func (q Query) Gavelspace() string { return q.gavelspace }
func (q Query) Limit() int         { return q.limit }
func (q Query) Offset() int        { return q.offset }
