package listfindings_test

import (
	"context"
	"sync"

	list "github.com/usegavel/gavel/core/application/casefile/listfindings"
)

type fakeFindingLister struct {
	mu    sync.Mutex
	items []list.FindingView
	total int
	err   error
}

func (f *fakeFindingLister) List(_ context.Context, _ string, _ list.Filters, _, _ int) ([]list.FindingView, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.items, f.total, nil
}
