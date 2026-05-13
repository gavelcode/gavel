package ingestfindings_test

import (
	"context"

	"github.com/usegavel/gavel/core/application/casefile/ingestfindings"
)

type fakeParser struct {
	parsed []ingestfindings.Parsed
	err    error
	calls  int
}

func (p *fakeParser) Parse(_ context.Context, _ []byte) ([]ingestfindings.Parsed, error) {
	p.calls++
	return p.parsed, p.err
}
