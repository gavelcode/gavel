package get_test

import (
	"context"
	"sync"

	"github.com/usegavel/gavel/core/application/project/get"
)

type fakeProjectGetter struct {
	mu     sync.Mutex
	result *get.ProjectDetail
	err    error
}

func (f *fakeProjectGetter) GetByID(_ context.Context, _ string) (*get.ProjectDetail, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}
