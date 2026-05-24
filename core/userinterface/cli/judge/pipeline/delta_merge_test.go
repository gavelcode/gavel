package pipeline_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/userinterface/cli/judge/pipeline"
)

func TestMergeDelta(t *testing.T) {
	findings := pipeline.Delta{
		NewCount:        3,
		FixedCount:      1,
		ExistingCount:   10,
		NewFingerprints: map[string]bool{"fp1": true},
		HasPrevious:     true,
	}
	arch := pipeline.Delta{
		NewViolationsCount:      2,
		FixedViolationsCount:    1,
		ExistingViolationsCount: 5,
		NewViolationIDs:         map[string]bool{"rule:a:b": true},
		HasArchPrevious:         true,
	}

	got := pipeline.MergeDelta(findings, arch)

	assert.Equal(t, 3, got.NewCount)
	assert.Equal(t, 1, got.FixedCount)
	assert.Equal(t, 10, got.ExistingCount)
	assert.True(t, got.HasPrevious)
	assert.Equal(t, 2, got.NewViolationsCount)
	assert.Equal(t, 1, got.FixedViolationsCount)
	assert.Equal(t, 5, got.ExistingViolationsCount)
	assert.True(t, got.HasArchPrevious)
	assert.True(t, got.NewViolationIDs["rule:a:b"])
}
