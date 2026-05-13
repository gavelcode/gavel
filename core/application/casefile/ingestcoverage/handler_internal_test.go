package ingestcoverage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubParser struct {
	result Parsed
	err    error
}

func (s *stubParser) Parse(_ context.Context, _ []byte) (Parsed, error) {
	return s.result, s.err
}

func TestExecuteNewEvidenceError(t *testing.T) {
	handler := &Handler{
		parsers: map[string]Parser{
			"lcov": &stubParser{result: Parsed{TotalLines: 100, CoveredLines: 80}},
		},
		now: func() time.Time { return time.Time{} },
	}

	cmd := Command{data: []byte("data"), format: "lcov", source: "bazel"}

	_, err := handler.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "build evidence")
}
