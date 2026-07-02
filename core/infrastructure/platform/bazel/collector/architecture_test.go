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

var _ collectevidence.ArchitectureCollector = (*collector.BazelArchitectureCollector)(nil)

func TestBazelArchitectureCollector_NoAspects_ReturnsNil(t *testing.T) {
	runner := &fakeAnalysisRunner{sarifFiles: map[string][][]byte{}}
	c := collector.NewBazelArchitectureCollector(runner)

	ev, docs, err := c.CollectViolations(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{"go": {"archtest"}})

	require.NoError(t, err)
	assert.Nil(t, ev)
	assert.Empty(t, docs)
}

func TestBazelArchitectureCollector_NoArchAspectsForUnknownLanguage(t *testing.T) {
	r := &fakeAnalysisRunner{sarifFiles: map[string][][]byte{}}
	c := collector.NewBazelArchitectureCollector(r)

	ev, docs, err := c.CollectViolations(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{})

	require.NoError(t, err)
	assert.Nil(t, ev)
	assert.Nil(t, docs)
}

func TestBazelArchitectureCollector_RunnerError(t *testing.T) {
	r := &fakeAnalysisRunner{err: fmt.Errorf("bazel failed")}
	c := collector.NewBazelArchitectureCollector(r)

	_, _, err := c.CollectViolations(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{"go": {"archtest"}})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "run archtest")
}

func TestBazelArchitectureCollector_InvalidSARIF(t *testing.T) {
	r := &fakeAnalysisRunner{sarifFiles: map[string][][]byte{"go_archtest_submission_aspect": {[]byte("{invalid}")}}}
	c := collector.NewBazelArchitectureCollector(r)

	_, _, err := c.CollectViolations(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{"go": {"archtest"}})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse archtest SARIF")
}

func TestBazelArchitectureCollector_ParsesViolations(t *testing.T) {
	archSarif := []byte(`{"runs":[{"results":[{"ruleId":"layer_violation","message":{"text":"forbidden"},"properties":{"sourcePkg":"api","targetPkg":"domain"}}]}]}`)
	runner := &fakeAnalysisRunner{sarifFiles: map[string][][]byte{"go_archtest_submission_aspect": {archSarif}}}
	c := collector.NewBazelArchitectureCollector(runner)

	ev, docs, err := c.CollectViolations(context.Background(), t.TempDir(), []string{"//pkg:lib"}, map[string][]string{"go": {"archtest"}})

	require.NoError(t, err)
	require.NotNil(t, ev)
	assert.Equal(t, "architecture", ev.Subtype)
	assert.Len(t, ev.Architecture.Violations, 1)
	assert.Len(t, docs, 1)
}
