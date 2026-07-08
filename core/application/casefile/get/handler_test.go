package get_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/shared/failure"

	"github.com/usegavel/gavel/core/application/casefile/get"
)

func TestNewQuery(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{name: "validQuery", id: "cf-1", wantErr: false},
		{name: "emptyID", id: "", wantErr: true},
		{name: "whitespaceID", id: "  ", wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			query, err := get.NewQuery("22222222-2222-2222-2222-222222222222", testCase.id)
			if testCase.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, get.ErrInvalidQuery)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.id, query.ID())
		})
	}
}

func TestNewHandlerPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { get.NewHandler(nil) })
}

func TestExecuteReturnsResult(t *testing.T) {
	now := time.Now()
	coverage := 92.3
	expected := &get.CaseFileDetail{
		ID:               "cf-1",
		ProjectID:        "proj-1",
		CommitSHA:        "abc123",
		Branch:           "main",
		StartedAt:        now,
		VerdictOutcome:   "pass",
		TotalFindings:    10,
		NewFindings:      2,
		ExistingFindings: 5,
		ResolvedFindings: 3,
		CoveragePercent:  &coverage,
		CreatedAt:        now,
		Evidences: []get.EvidenceSummary{
			{
				ID:          "ev-1",
				Subtype:     "lint",
				Source:      "pmd",
				CollectedAt: now,
			},
		},
		Rulings: []get.RulingView{
			{
				Subtype: "new_findings",
				Passed:  true,
				Detail:  "0 new findings",
			},
		},
	}
	fake := &fakeCaseFileGetter{result: expected}
	h := get.NewHandler(fake)

	query, err := get.NewQuery("22222222-2222-2222-2222-222222222222", "cf-1")
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), query)

	require.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.ProjectID, result.ProjectID)
	assert.Equal(t, expected.VerdictOutcome, result.VerdictOutcome)
	assert.Len(t, result.Evidences, 1)
	assert.Equal(t, "ev-1", result.Evidences[0].ID)
	assert.Len(t, result.Rulings, 1)
	assert.True(t, result.Rulings[0].Passed)
}

func TestExecutePropagatesError(t *testing.T) {
	queryErr := errors.New("not found")
	fake := &fakeCaseFileGetter{err: queryErr}
	h := get.NewHandler(fake)

	query, err := get.NewQuery("22222222-2222-2222-2222-222222222222", "cf-1")
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), query)

	assert.ErrorIs(t, err, queryErr)
}

func TestErrInvalidQueryIsClassifiedAsValidation(t *testing.T) {
	_, err := get.NewQuery("22222222-2222-2222-2222-222222222222", "")
	require.Error(t, err)
	assert.Equal(t, failure.Validation, failure.Of(err))
}
