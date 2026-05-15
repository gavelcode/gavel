package search_test

import (
	"context"
	"sync"

	"github.com/usegavel/gavel/core/application/supporting/search"
)

type fakeSearchQuery struct {
	mu      sync.Mutex
	results []search.SearchResult
	err     error
}

func (f *fakeSearchQuery) Search(_ context.Context, _ string, _ int) ([]search.SearchResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	return f.results, nil
}
