package ingestcoverage_test

import (
	"context"

	"github.com/usegavel/gavel/core/application/casefile/ingestcoverage"
)

type fakeParser struct {
	parsed ingestcoverage.Parsed
	err    error
	calls  int
	data   []byte
}

func (p *fakeParser) Parse(_ context.Context, data []byte) (ingestcoverage.Parsed, error) {
	p.calls++
	p.data = data
	return p.parsed, p.err
}
