package loadgavelspace

import (
	"fmt"
	"strings"
)

type Query struct {
	configPath    string
	workspace     string
	projectFilter string
}

func NewQuery(configPath string, opts ...QueryOption) (Query, error) {
	if strings.TrimSpace(configPath) == "" {
		return Query{}, fmt.Errorf("%w: configPath must not be empty", ErrInvalidQuery)
	}
	q := Query{configPath: configPath}
	for _, opt := range opts {
		opt(&q)
	}
	return q, nil
}

type QueryOption func(*Query)

func WithWorkspace(workspace string) QueryOption {
	return func(q *Query) { q.workspace = workspace }
}

func WithProjectFilter(name string) QueryOption {
	return func(q *Query) { q.projectFilter = name }
}

func (q Query) ConfigPath() string    { return q.configPath }
func (q Query) Workspace() string     { return q.workspace }
func (q Query) ProjectFilter() string { return q.projectFilter }
