package get_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/application/pleading/get"
	"github.com/usegavel/gavel/core/domain/shared/failure"
)

func TestNewQuery(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{name: "validQuery", id: "pr-1", wantErr: false},
		{name: "emptyID", id: "", wantErr: true},
		{name: "whitespaceID", id: "  ", wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			query, err := get.NewQuery("22222222-2222-2222-2222-222222222222", testCase.id)
			if testCase.wantErr {
				assert.Error(t, err)
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
	expected := &get.PleadingDetail{
		ID:           "pr-1",
		ProjectID:    "proj-1",
		Number:       42,
		Title:        "Fix bug",
		Petitioner:   "dev",
		SourceBranch: "fix/bug",
		TargetBranch: "main",
		CommitSHA:    "abc123",
		Status:       "open",
		GateResult:   &get.GateResult{Passed: true, Conditions: []get.GateCondition{{Label: "coverage", Operator: ">=", Value: "80", Threshold: "70", Passed: true}}},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	fake := &fakePleadingGetter{result: expected}
	h := get.NewHandler(fake)

	query, err := get.NewQuery("22222222-2222-2222-2222-222222222222", "pr-1")
	require.NoError(t, err)

	result, err := h.Execute(context.Background(), query)

	require.NoError(t, err)
	assert.Equal(t, "pr-1", result.ID)
	assert.Equal(t, "proj-1", result.ProjectID)
	assert.Equal(t, 42, result.Number)
	assert.Equal(t, "Fix bug", result.Title)
	assert.Equal(t, "dev", result.Petitioner)
	assert.True(t, result.GateResult.Passed)
}

func TestExecutePropagatesError(t *testing.T) {
	queryErr := errors.New("connection failed")
	fake := &fakePleadingGetter{err: queryErr}
	h := get.NewHandler(fake)

	query, err := get.NewQuery("22222222-2222-2222-2222-222222222222", "pr-1")
	require.NoError(t, err)

	_, err = h.Execute(context.Background(), query)

	assert.ErrorIs(t, err, queryErr)
}

func TestErrInvalidQueryIsClassifiedAsValidation(t *testing.T) {
	_, err := get.NewQuery("22222222-2222-2222-2222-222222222222", "")
	require.Error(t, err)
	assert.Equal(t, failure.Validation, failure.Of(err))
}
