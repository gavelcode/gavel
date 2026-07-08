package finalize_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/usegavel/gavel/core/application/casefile/finalize"
)

func TestNewCommandRejectsEmptyCaseFileID(t *testing.T) {
	_, err := finalize.NewCommand(testTenant.String(), "")
	assert.ErrorIs(t, err, finalize.ErrInvalidCommand)
}

func TestNewCommandRejectsPrecomputedVerdictZeroEvaluatedAt(t *testing.T) {
	_, err := finalize.NewCommand(testTenant.String(), "some-id",
		finalize.WithPrecomputedVerdict(finalize.PrecomputedVerdict{
			Outcome:     "pass",
			EvaluatedAt: time.Time{},
		}),
	)
	assert.ErrorIs(t, err, finalize.ErrInvalidCommand)
}
