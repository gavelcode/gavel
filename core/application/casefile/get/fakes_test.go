package get_test

import (
	"context"
	"sync"

	"github.com/usegavel/gavel/core/application/casefile/get"
)

type fakeCaseFileGetter struct {
	mu     sync.Mutex
	result *get.CaseFileDetail
	err    error
}

func (f *fakeCaseFileGetter) GetByID(_ context.Context, _, _ string) (*get.CaseFileDetail, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}
