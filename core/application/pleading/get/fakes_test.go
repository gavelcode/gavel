package get_test

import (
	"context"
	"sync"

	"github.com/usegavel/gavel/core/application/pleading/get"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

type fakePleadingGetter struct {
	mu     sync.Mutex
	result *get.PleadingDetail
	err    error
}

func (f *fakePleadingGetter) GetByID(_ context.Context, _ tenant.TenantID, _ string) (*get.PleadingDetail, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}
