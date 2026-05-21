package runner

import (
	"context"
	"fmt"
)

type fakeCall struct {
	Dir  string
	Name string
	Args []string
}

type fakeResult struct {
	Stdout []byte
	Stderr []byte
	Err    error
}

type fakeRunner struct {
	calls   []fakeCall
	results []fakeResult
	idx     int
	runHook func(dir string, args []string)
}

func (f *fakeRunner) Run(_ context.Context, dir, name string, args ...string) ([]byte, []byte, error) {
	f.calls = append(f.calls, fakeCall{Dir: dir, Name: name, Args: args})
	if f.runHook != nil {
		f.runHook(dir, args)
	}
	if f.idx >= len(f.results) {
		return nil, nil, fmt.Errorf("fakeRunner: unexpected call #%d to %s %v", f.idx+1, name, args)
	}
	r := f.results[f.idx]
	f.idx++
	return r.Stdout, r.Stderr, r.Err
}
