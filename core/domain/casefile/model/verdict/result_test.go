package verdict_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/verdict"
)

func TestComposeVerdict(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		rulings         []verdict.Ruling
		expectedOutcome string
	}{
		{
			name:            "all rulings pass yields OutcomePass",
			rulings:         []verdict.Ruling{passRuling(t, evidence.SubtypeCodeQuality)},
			expectedOutcome: "pass",
		},
		{
			name: "ruling fails yields OutcomeFail",
			rulings: []verdict.Ruling{
				failRuling(t, evidence.SubtypeCodeQuality, "5 errors (max 0)"),
			},
			expectedOutcome: "fail",
		},
		{
			name:            "empty rulings yields OutcomePass",
			rulings:         nil,
			expectedOutcome: "pass",
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			verdict, err := verdict.Compose(tcase.rulings, now)

			require.NoError(t, err)
			assert.Equal(t, tcase.expectedOutcome, verdict.Outcome().String())
			assert.Equal(t, now, verdict.EvaluatedAt())
		})
	}
}

func TestReconstituteVerdict(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	rulings := []verdict.Ruling{
		passRuling(t, evidence.SubtypeCodeQuality),
	}

	verdict, err := verdict.ReconstituteResult("pass", rulings, now)

	require.NoError(t, err)
	assert.Equal(t, "pass", verdict.Outcome().String())
	assert.Equal(t, now, verdict.EvaluatedAt())
	require.Len(t, verdict.Rulings(), 1)
}

func TestReconstituteVerdictInvalidOutcomeRejected(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)

	_, err := verdict.ReconstituteResult("invalid", nil, now)

	require.Error(t, err)
}

func TestComposeVerdictRejectsZeroEvaluatedAt(t *testing.T) {
	_, err := verdict.Compose(nil, time.Time{})

	require.Error(t, err, "ComposeVerdict must reject a zero evaluatedAt — matches NewCaseFile / NewEvidence invariant style")
}

func TestReconstituteVerdictRejectsZeroEvaluatedAt(t *testing.T) {
	_, err := verdict.ReconstituteResult("pass", nil, time.Time{})

	require.Error(t, err, "ReconstituteVerdict must reject a zero evaluatedAt")
}

func TestVerdictRulingsDefensiveCopy(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	rulings := []verdict.Ruling{
		passRuling(t, evidence.SubtypeCodeQuality),
	}

	verdict, err := verdict.Compose(rulings, now)
	require.NoError(t, err)

	returned := verdict.Rulings()
	returned[0] = failRuling(t, evidence.SubtypeSAST, "mutated")

	assert.Equal(t, evidence.SubtypeCodeQuality, verdict.Rulings()[0].Subtype())
}

func passRuling(_ *testing.T, subtype evidence.Subtype) verdict.Ruling {
	return verdict.NewRuling(subtype, true, "")
}

func failRuling(_ *testing.T, subtype evidence.Subtype, detail string) verdict.Ruling {
	return verdict.NewRuling(subtype, false, detail)
}
