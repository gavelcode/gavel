package get_test

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/usegavel/gavel/core/application/project/get"
	"github.com/usegavel/gavel/core/domain/iam/model/tenant"
)

var testTenant = tenant.NewTenantID(uuid.MustParse("22222222-2222-2222-2222-222222222222"))

type fakeProjectGetter struct {
	mu     sync.Mutex
	result *get.ProjectDetail
	err    error
}

func (f *fakeProjectGetter) GetByID(_ context.Context, _ tenant.TenantID, _ string) (*get.ProjectDetail, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}
