package pipeline_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/userinterface/cli/judge/pipeline"
)

func TestLocalMode_FullPipeline(t *testing.T) {
	fix := newLocalFixture(t)
	proj := mustSeedProject(t, fix.projRepo, "full-pipeline")

	collected := buildCollectedEvidence(t, []string{"fp-1", "fp-2"}, 85.0)

	result, err := pipeline.RunLocal(
		context.Background(),
		fix.deps,
		testWorkspace,
		collected,
		proj.ID().String(),
		proj.Name(),
		testCommitSHA,
		testBranch,
		time.Now().UTC(),
		pipeline.Options{},
	)
	require.NoError(t, err)

	assert.Equal(t, "pass", result.Verdict)
	assert.Equal(t, 2, result.FindingsCount)
	assert.InDelta(t, 85.0, result.CoveragePercent, 0.1)
}

func TestLocalMode_FailingGate(t *testing.T) {
	fix := newLocalFixture(t)
	gate := mustBuildZeroToleranceGate(t)
	proj := mustSeedProject(t, fix.projRepo, "failing-gate", withGate(gate))

	collected := buildCollectedEvidence(t, []string{"fp-1", "fp-2"}, 80.0)

	result, err := pipeline.RunLocal(
		context.Background(),
		fix.deps,
		testWorkspace,
		collected,
		proj.ID().String(),
		proj.Name(),
		testCommitSHA,
		testBranch,
		time.Now().UTC(),
		pipeline.Options{},
	)
	require.NoError(t, err)

	assert.Equal(t, "fail", result.Verdict)
	assert.NotEmpty(t, result.Rulings)
}

func TestLocalMode_BaselineDelta(t *testing.T) {
	fix := newLocalFixture(t)
	proj := mustSeedProject(t, fix.projRepo, "baseline-delta")

	firstCollected := buildCollectedEvidence(t, []string{"fp-1", "fp-2"}, 80.0)
	_, err := pipeline.RunLocal(
		context.Background(),
		fix.deps,
		testWorkspace,
		firstCollected,
		proj.ID().String(),
		proj.Name(),
		testCommitSHA,
		testBranch,
		time.Now().UTC(),
		pipeline.Options{},
	)
	require.NoError(t, err)

	secondCollected := buildCollectedEvidence(t, []string{"fp-2", "fp-3"}, 80.0)
	result, err := pipeline.RunLocal(
		context.Background(),
		fix.deps,
		testWorkspace,
		secondCollected,
		proj.ID().String(),
		proj.Name(),
		testCommitSHA,
		testBranch,
		time.Now().UTC(),
		pipeline.Options{},
	)
	require.NoError(t, err)

	assert.Equal(t, 1, result.Delta.ExistingCount)
	assert.Equal(t, 1, result.Delta.NewCount)
	assert.Equal(t, 1, result.Delta.FixedCount)
	assert.True(t, result.Delta.HasPrevious)
}

func TestLocalMode_FirstRunHasNoPrevious(t *testing.T) {
	fix := newLocalFixture(t)
	proj := mustSeedProject(t, fix.projRepo, "first-run")

	collected := buildCollectedEvidence(t, []string{"fp-1"}, 90.0)

	result, err := pipeline.RunLocal(
		context.Background(),
		fix.deps,
		testWorkspace,
		collected,
		proj.ID().String(),
		proj.Name(),
		testCommitSHA,
		testBranch,
		time.Now().UTC(),
		pipeline.Options{},
	)
	require.NoError(t, err)

	assert.False(t, result.Delta.HasPrevious)
	assert.True(t, result.FirstRun)
}

func TestLocalMode_AbsoluteMode(t *testing.T) {
	fix := newLocalFixture(t)
	gate := mustBuildZeroToleranceGate(t)
	proj := mustSeedProject(t, fix.projRepo, "absolute-mode",
		withGate(gate),
		withBaseline(testBranch, []string{"fp-existing"}),
	)

	collected := buildCollectedEvidence(t, []string{"fp-existing"}, 80.0)

	result, err := pipeline.RunLocal(
		context.Background(),
		fix.deps,
		testWorkspace,
		collected,
		proj.ID().String(),
		proj.Name(),
		testCommitSHA,
		testBranch,
		time.Now().UTC(),
		pipeline.Options{Absolute: true},
	)
	require.NoError(t, err)

	assert.Equal(t, "fail", result.Verdict)
	assert.Equal(t, 0, result.Delta.NewCount)
	assert.Equal(t, 0, result.Delta.FixedCount)
	assert.Equal(t, 0, result.Delta.ExistingCount)
}
