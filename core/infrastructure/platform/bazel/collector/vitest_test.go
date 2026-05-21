package collector_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/collector"
)

func TestVitestCollectorEmptyTargets(t *testing.T) {
	c := collector.NewVitestCoverageCollector(slog.Default())

	data, err := c.CollectCoverage(context.Background(), t.TempDir(), nil, nil)

	require.NoError(t, err)
	assert.Nil(t, data)
}
