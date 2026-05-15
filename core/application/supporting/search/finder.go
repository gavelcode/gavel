package search

import "context"

type Finder interface {
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
}
