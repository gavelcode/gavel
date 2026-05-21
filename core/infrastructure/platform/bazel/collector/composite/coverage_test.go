package composite_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/casefile/collectevidence"
	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/collector/composite"
)

var _ collectevidence.CoverageCollector = (*composite.CoverageCollector)(nil)

type fakeCoverageCollector struct {
	data []byte
	err  error
}

func (f *fakeCoverageCollector) CollectCoverage(_ context.Context, _ string, _ []string, _ []string) ([]byte, error) {
	return f.data, f.err
}

func TestCompositePrimaryReturnsData(t *testing.T) {
	lcov := []byte("SF:main.go\nDA:1,1\nLF:1\nLH:1\nend_of_record\n")
	primary := &fakeCoverageCollector{data: lcov}
	fallback := &fakeCoverageCollector{data: []byte("should not be used")}

	collector := composite.NewCoverageCollector(primary, fallback)

	data, err := collector.CollectCoverage(context.Background(), "/ws", []string{"//app/..."}, []string{"typescript"})

	require.NoError(t, err)
	assert.Equal(t, lcov, data)
}

func TestCompositeFallbackWhenPrimaryEmpty(t *testing.T) {
	vitestLCOV := []byte("SF:src/app.tsx\nDA:1,1\nLF:1\nLH:1\nend_of_record\n")
	primary := &fakeCoverageCollector{data: nil}
	fallback := &fakeCoverageCollector{data: vitestLCOV}

	collector := composite.NewCoverageCollector(primary, fallback)

	data, err := collector.CollectCoverage(context.Background(), "/ws", []string{"//app/..."}, []string{"typescript"})

	require.NoError(t, err)
	assert.Equal(t, vitestLCOV, data)
}

func TestCompositeNoFallbackWithoutTypeScript(t *testing.T) {
	primary := &fakeCoverageCollector{data: nil}
	fallback := &fakeCoverageCollector{data: []byte("should not be used")}

	collector := composite.NewCoverageCollector(primary, fallback)

	data, err := collector.CollectCoverage(context.Background(), "/ws", []string{"//core/..."}, []string{"go"})

	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestCompositePrimaryError(t *testing.T) {
	primary := &fakeCoverageCollector{err: fmt.Errorf("bazel failed")}
	fallback := &fakeCoverageCollector{data: []byte("fallback")}

	coll := composite.NewCoverageCollector(primary, fallback)

	_, err := coll.CollectCoverage(context.Background(), "/ws", []string{"//app/..."}, []string{"typescript"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bazel failed")
}

func TestCompositeNilFallback(t *testing.T) {
	primary := &fakeCoverageCollector{data: nil}

	coll := composite.NewCoverageCollector(primary, nil)

	data, err := coll.CollectCoverage(context.Background(), "/ws", []string{"//app/..."}, []string{"typescript"})

	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestCompositeFallbackError_ReturnsPrimaryData(t *testing.T) {
	emptyCov := []byte("SF:test.py\nLF:0\nLH:0\nend_of_record\n")
	primary := &fakeCoverageCollector{data: emptyCov}
	fallback := &fakeCoverageCollector{err: fmt.Errorf("vitest failed")}

	coll := composite.NewCoverageCollector(primary, fallback)

	data, err := coll.CollectCoverage(context.Background(), "/ws", []string{"//app/..."}, []string{"typescript"})

	require.NoError(t, err)
	assert.Equal(t, emptyCov, data)
}

func TestCompositeFallbackReturnsNil_ReturnsPrimaryData(t *testing.T) {
	emptyCov := []byte("SF:test.py\nLF:0\nLH:0\nend_of_record\n")
	primary := &fakeCoverageCollector{data: emptyCov}
	fallback := &fakeCoverageCollector{data: nil}

	coll := composite.NewCoverageCollector(primary, fallback)

	data, err := coll.CollectCoverage(context.Background(), "/ws", []string{"//app/..."}, []string{"typescript"})

	require.NoError(t, err)
	assert.Equal(t, emptyCov, data)
}

func TestCompositeFallbackOnZeroCoverage(t *testing.T) {
	emptyCov := []byte("SF:test.py\nLF:0\nLH:0\nend_of_record\n")
	vitestLCOV := []byte("SF:src/app.tsx\nDA:1,1\nLF:1\nLH:1\nend_of_record\n")
	primary := &fakeCoverageCollector{data: emptyCov}
	fallback := &fakeCoverageCollector{data: vitestLCOV}

	collector := composite.NewCoverageCollector(primary, fallback)

	data, err := collector.CollectCoverage(context.Background(), "/ws", []string{"//app/..."}, []string{"typescript"})

	require.NoError(t, err)
	assert.Equal(t, vitestLCOV, data)
}
