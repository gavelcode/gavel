package list_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/shared/failure"

	"github.com/usegavel/gavel/core/application/pleading/list"
)

func TestNewQuery(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		status    string
		limit     int
		offset    int
		wantErr   bool
	}{
		{name: "validQuery", projectID: "proj-1", status: "open", limit: 10, offset: 0, wantErr: false},
		{name: "emptyProjectIDIsValid", projectID: "", status: "open", limit: 10, offset: 0, wantErr: false},
		{name: "emptyStatusIsValid", projectID: "proj-1", status: "", limit: 10, offset: 0, wantErr: false},
		{name: "zeroLimit", projectID: "proj-1", status: "open", limit: 0, offset: 0, wantErr: true},
		{name: "negativeLimit", projectID: "proj-1", status: "open", limit: -1, offset: 0, wantErr: true},
		{name: "negativeOffset", projectID: "proj-1", status: "open", limit: 10, offset: -1, wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			query, err := list.NewQuery(testCase.projectID, testCase.status, "", testCase.limit, testCase.offset)
			if testCase.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.projectID, query.ProjectID())
			assert.Equal(t, testCase.status, query.Status())
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
	expected := []list.PleadingSummary{
		{
			ID:           "pr-1",
			ProjectID:    "proj-1",
			Number:       42,
			Title:        "Fix bug",
			Petitioner:   "dev",
			SourceBranch: "fix/bug",
			TargetBranch: "main",
			CommitSHA:    "abc123",
			Status:       "open",
			GateResult:   &list.GateResult{Passed: true, Conditions: []list.GateCondition{{Label: "coverage", Operator: ">=", Value: "80", Threshold: "70", Passed: true}}},
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
	fake := &fakePleadingLister{items: expected, total: 1}
	h := list.NewHandler(fake)

	query, err := list.NewQuery("proj-1", "open", "", 10, 0)
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), query)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "pr-1", result.Items[0].ID)
	assert.Equal(t, "proj-1", result.Items[0].ProjectID)
	assert.Equal(t, 42, result.Items[0].Number)
	assert.True(t, result.Items[0].GateResult.Passed)
}

func TestExecutePropagatesError(t *testing.T) {
	queryErr := errors.New("connection failed")
	fake := &fakePleadingLister{err: queryErr}
	h := list.NewHandler(fake)

	query, err := list.NewQuery("proj-1", "open", "", 10, 0)
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), query)

	assert.ErrorIs(t, err, queryErr)
}

func TestErrInvalidQueryIsClassifiedAsValidation(t *testing.T) {
	_, err := list.NewQuery("", "", "", 0, 0)
	require.Error(t, err)
	assert.Equal(t, failure.Validation, failure.Of(err))
}
