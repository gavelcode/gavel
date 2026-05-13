package list_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/shared/failure"

	"github.com/usegavel/gavel/core/application/casefile/list"
)

func TestNewQuery(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		limit     int
		offset    int
		wantErr   bool
	}{
		{name: "validQuery", projectID: "proj-1", limit: 10, offset: 0, wantErr: false},
		{name: "emptyProjectID", projectID: "", limit: 10, offset: 0, wantErr: true},
		{name: "whitespaceProjectID", projectID: "  ", limit: 10, offset: 0, wantErr: true},
		{name: "zeroLimit", projectID: "proj-1", limit: 0, offset: 0, wantErr: true},
		{name: "negativeLimit", projectID: "proj-1", limit: -1, offset: 0, wantErr: true},
		{name: "negativeOffset", projectID: "proj-1", limit: 10, offset: -1, wantErr: true},
		{name: "validWithOffset", projectID: "proj-1", limit: 5, offset: 20, wantErr: false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			query, err := list.NewQuery(testCase.projectID, "", testCase.limit, testCase.offset)
			if testCase.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, list.ErrInvalidQuery)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.projectID, query.ProjectID())
			assert.Equal(t, testCase.limit, query.Limit())
			assert.Equal(t, testCase.offset, query.Offset())
		})
	}
}

func TestNewHandlerPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { list.NewHandler(nil) })
}

func TestExecuteReturnsResult(t *testing.T) {
	now := time.Now()
	coverage := 85.5
	expected := []list.CaseFileSummary{
		{
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
		},
	}
	fake := &fakeCaseFileLister{items: expected, total: 42}
	h := list.NewHandler(fake)

	query, err := list.NewQuery("proj-1", "", 10, 0)
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), query)

	require.NoError(t, err)
	assert.Equal(t, expected, result.Items)
	assert.Equal(t, 42, result.Total)
}

func TestExecutePropagatesError(t *testing.T) {
	queryErr := errors.New("connection failed")
	fake := &fakeCaseFileLister{err: queryErr}
	h := list.NewHandler(fake)

	query, err := list.NewQuery("proj-1", "", 10, 0)
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), query)

	assert.ErrorIs(t, err, queryErr)
}

func TestErrInvalidQueryIsClassifiedAsValidation(t *testing.T) {
	_, err := list.NewQuery("", "", 10, 0)
	require.Error(t, err)
	assert.Equal(t, failure.Validation, failure.Of(err))
}
