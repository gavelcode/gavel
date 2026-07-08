package list_test

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/usegavel/gavel/core/application/project/list"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

var testTenant = tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))

type fakeProjectLister struct {
	mu    sync.Mutex
	items []list.ProjectSummary
	total int
	err   error
}

func (f *fakeProjectLister) List(_ context.Context, _ tenant.TenantID, _, _ int) ([]list.ProjectSummary, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.items, f.total, nil
}
