package list_test

import (
	"context"
	"sync"

	"github.com/usegavel/gavel/core/application/casefile/list"
)

type fakeCaseFileLister struct {
	mu    sync.Mutex
	items []list.CaseFileSummary
	total int
	err   error
}

func (f *fakeCaseFileLister) ListByProject(_ context.Context, _, _, _ string, _, _ int) ([]list.CaseFileSummary, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.items, f.total, nil
}
