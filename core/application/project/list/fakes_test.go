package list_test

import (
	"context"
	"sync"

	"github.com/usegavel/gavel/core/application/project/list"
)

type fakeProjectLister struct {
	mu    sync.Mutex
	items []list.ProjectSummary
	total int
	err   error
}

func (f *fakeProjectLister) List(_ context.Context, _, _ int) ([]list.ProjectSummary, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.items, f.total, nil
}
