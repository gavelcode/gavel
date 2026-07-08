package list_test

import (
	"context"
	"sync"

	"github.com/usegavel/gavel/core/application/pleading/list"
)

type fakePleadingLister struct {
	mu    sync.Mutex
	items []list.PleadingSummary
	total int
	err   error
}

func (f *fakePleadingLister) ListByProject(_ context.Context, _, _, _, _ string, _, _ int) ([]list.PleadingSummary, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.items, f.total, nil
}
