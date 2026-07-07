package get_test

import (
	"context"
	"sync"

	"github.com/usegavel/gavel/core/application/gavelspace/get"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

type fakeGavelspaceGetter struct {
	mu     sync.Mutex
	result *get.GavelspaceDetail
	err    error
}

func (f *fakeGavelspaceGetter) GetByName(_ context.Context, _ tenant.TenantID, _ string) (*get.GavelspaceDetail, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

const testTenant = "22222222-2222-2222-2222-222222222222"
