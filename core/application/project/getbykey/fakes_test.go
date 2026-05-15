package getbykey_test

import (
	"context"
	"sync"

	"github.com/usegavel/gavel/core/application/project/getbykey"
)

type fakeProjectKeyGetter struct {
	mu     sync.Mutex
	result *getbykey.ProjectDetail
	err    error
}

func (f *fakeProjectKeyGetter) GetByKey(_ context.Context, _ string) (*getbykey.ProjectDetail, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}
