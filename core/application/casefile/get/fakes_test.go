package get_test

import (
	"context"
	"sync"

	"github.com/usegavel/gavel/core/application/casefile/get"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

type fakeCaseFileGetter struct {
	mu     sync.Mutex
	result *get.CaseFileDetail
	err    error
}

func (f *fakeCaseFileGetter) GetByID(_ context.Context, _ tenant.TenantID, _ string) (*get.CaseFileDetail, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}
