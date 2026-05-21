package collector_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/collector"
)

var _ collectevidence.CoverageCollector = (*collector.BazelCoverageCollector)(nil)

func TestBazelCoverageCollector_ReturnsCoverageData(t *testing.T) {
	lcov := []byte("SF:main.go\nDA:1,1\nend_of_record\n")
	runner := &fakeAnalysisRunner{coverageData: lcov}
	c := collector.NewBazelCoverageCollector(runner)

	data, err := c.CollectCoverage(context.Background(), t.TempDir(), []string{"//pkg:lib"}, nil)

	require.NoError(t, err)
	assert.Equal(t, lcov, data)
}

func TestBazelCoverageCollector_RunnerError(t *testing.T) {
	r := &fakeAnalysisRunner{err: fmt.Errorf("bazel coverage failed")}
	c := collector.NewBazelCoverageCollector(r)

	_, err := c.CollectCoverage(context.Background(), t.TempDir(), []string{"//pkg:lib"}, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "run coverage")
}

func TestBazelCoverageCollector_NilWhenNoCoverage(t *testing.T) {
	runner := &fakeAnalysisRunner{}
	c := collector.NewBazelCoverageCollector(runner)

	data, err := c.CollectCoverage(context.Background(), t.TempDir(), []string{"//pkg:lib"}, nil)

	require.NoError(t, err)
	assert.Nil(t, data)
}
