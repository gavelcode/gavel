package list_test

import (
	"context"
	"sync"

	"github.com/usegavel/gavel/core/application/gavelspace/list"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

type fakeGavelspaceLister struct {
	mu    sync.Mutex
	items []list.GavelspaceSummary
	total int
	err   error
}

func (f *fakeGavelspaceLister) List(_ context.Context, _ tenant.TenantID, _, _ int) ([]list.GavelspaceSummary, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.items, f.total, nil
}

const testTenant = "22222222-2222-2222-2222-222222222222"
